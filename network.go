package brain

import (
    "fmt"
    "math/rand"
    "os"
    "strconv"
    "encoding/json"
)

/*
    & is "address of"
    * is "value at address"
*/

// http://www.scientificamerican.com/article/ask-the-brains-aug-08/

type ConnInfo struct {
    Excitatory bool  `json:"excitatory"`
    Strength float64 `json:"strength"`
}

type Connection struct {
    To map[*Node]*ConnInfo  `json:"to"`
    HoldingVal int          `json:"holding"`
    Center [3]int           `json:"center"` // todo - maybe float and then round when generating?
}

type Node struct {
    Value int                                  `json:"value"`
    OutgoingConnection *Connection             `json:"-"`  //which node to send to
    IncomingConnections map[*Node]*Connection  `json:"-"`  //which nodes to read from
    Position [3]int                            `json:"position"`
    Id string                                  `json:"id"` //"x|y|z"
}

// let's say y=0 is the front of the "brain"
type Network struct {
    Nodes [][][]*Node             `json:"nodes"`
    LeftHemisphere [][][]*Node    `json:"-"`
    RightHemisphere [][][]*Node   `json:"-"`
    Dimensions [3]int             `json:"-"`
    Sensors map[string]*Sensor    `json:"-"`
    Outputs map[string]*Output    `json:"-"`
    Frames int                    `json:"-"`
}

func (c Connection) String() string {
    jsonRep, _ := json.MarshalIndent(c, "", "    ")
    return string(jsonRep)
}

func (n *Node) Update() {
    // calculate if it's going to fire or not - calculate sum and then set to 1, 0
    // base sum on excitatory/inhibiting
    var sum float64

    // first calculate sum
    for _, conn := range n.IncomingConnections {
        if conn.To[n].Excitatory {
            sum = sum + (float64(conn.HoldingVal) * conn.To[n].Strength)
        } else {
            sum = sum - (float64(conn.HoldingVal) * conn.To[n].Strength)
        }
    }

    if sum >= FIRING_THRESHOLD {
        n.Value = 1
    }

    // then, based on whether it fired, prune/strengthen connections
    // magic numbers.
    // additive or multiplicative?
    // maybe as a fraction/percent of distance from some constant (0.5?  0.75)
    for from, conn := range n.IncomingConnections {
        // adjusting
        if conn.HoldingVal == n.Value { // nodes worked in conjunction...
            if n.Value == 1 { // and both fired (if neither fired, decay a little bit)
                conn.To[n].Strength += CONN_WEIGHT_INCREASE
            } else {
                // decay a little bit
                conn.To[n].Strength -= CONN_WEIGHT_DECAY
            }
        } else {
            // the nodes didn't fire together
            conn.To[n].Strength -= CONN_WEIGHT_DECREASE
        }

        // pruning
        if conn.To[n].Strength > MAX_CONN_WEIGHT {
            conn.To[n].Strength = MAX_CONN_WEIGHT
        }
        if conn.To[n].Strength < MIN_CONN_WEIGHT {
            delete(conn.To, n)
            delete(n.IncomingConnections, from)
        }
    }

}

func RandFloat(min, max float64) float64 {
    randFloat := rand.Float64()
    diff := max - min
    r := randFloat * diff
    return min + r
}

func (net *Network) AddConnections(node *Node) {
    center := node.OutgoingConnection.Center
    possibleExtensions := []*Node{}
    numPossible := rand.Intn(15 - 5) + 5 // magic - 10 to 15
    stDev := DYNAMIC_SYNAPSE_PROB_SPHERE
    for i := 0; i < numPossible; i++ {
        potCenter := node.Position
        for potCenter == node.Position {
            potX := int(rand.NormFloat64() * stDev) + center[0]
            potY := int(rand.NormFloat64() * stDev) + center[1]
            potZ := int(rand.NormFloat64() * stDev) + center[2]
            for potX < 0 || potX >= (net.Dimensions[0] * 2) {
                potX = int(rand.NormFloat64() * stDev) + center[0]
            }
            for potY < 0 || potY >= net.Dimensions[1] {
                potY = int(rand.NormFloat64() * stDev) + center[1]
            }
            for potZ < 0 || potZ >= net.Dimensions[2] {
                potZ = int(rand.NormFloat64() * stDev) + center[2]
            }
            potCenter = [3]int{potX, potY, potZ}
        }
        possibleExtensions = append(possibleExtensions, net.FindNode(potCenter))
    }
    for _, potNode := range possibleExtensions {
        _, exists := node.OutgoingConnection.To[potNode]
        if potNode.Value != 0 && !exists {
            excitatory := false
            if rand.Intn(INVERSE_INHIBITORY_PROB) != 0 {
                // https://www.quora.com/Are-there-more-excitatory-neurons-or-inhibitory-neurons-in-the-brain-Why
                excitatory = true
            }
            node.OutgoingConnection.To[potNode] = &ConnInfo{
                Strength: RandFloat(0.25, 0.75),
                Excitatory: excitatory,
            }
            potNode.IncomingConnections[node] = node.OutgoingConnection
        } 
    }
}

func (node *Node) UpdateOutgoingCenter() {
    x := 0
    y := 0
    z := 0
    numOutgoing := len(node.OutgoingConnection.To)
    if numOutgoing < 0 {
        for to := range node.OutgoingConnection.To {
            x += to.Position[0]
            y += to.Position[1]
            z += to.Position[2]
        }
        x = x / numOutgoing
        y = y / numOutgoing
        z = z / numOutgoing
        node.OutgoingConnection.Center = [3]int{x, y, z}
    }
}

// let's see which one causes the most overhead...
// or it might just be all of them
func (net *Network) Cycle() {
    // fake concurrency
    // first, set all the connections based on their nodes

    net.ForEachNode(func(node *Node, pos [3]int) {
        if node.Value != 0 {
            net.AddConnections(node)
        }
        node.OutgoingConnection.HoldingVal = node.Value
        node.Value = 0
    })
    

    // then set all the nodes based on connections
    net.ForEachNode(func(node *Node, pos [3]int) {
        node.Update()
        node.UpdateOutgoingCenter()
    })


    // also update nodes that receive sensory information
    for _, output := range net.Outputs {
        output.Value = output.Out(output.Nodes)
    }

    for _, sensor := range net.Sensors {
        sensor.In(sensor.Nodes, sensor.Influences)
    }
    
    net.Frames++
}

////////////////////////////////////////////////////////////////////
//                                                                //
//                  Mostly generation stuff past here             //
//                                                                //
////////////////////////////////////////////////////////////////////

func (n Node) String() string {
    jsonRep, _ := json.MarshalIndent(n, "", "    ")
    return string(jsonRep)
}

// is this still needed?
// for cool viz, sure
func (net *Network) RandomizeValues(probOn float32) {
    net.ForEachNode(func(node *Node, pos [3]int) {
        temp := rand.Float32()
        if temp < probOn {
            node.Value = 1
        } else {
            node.Value = 0
        }
    })
}

func NodeExistsIn(node *Node, nodes []*Node) bool {
    for _, potNode := range nodes {
        if (node == potNode) {
            return true
        }
    }
    return false
}

func (net *Network) ConnectHemispheres() {
    net.ForEachNode(func(node *Node, pos [3]int) {
        // fmt.Println(node.OutgoingConnection.Center)
        centralConnNode := net.FindNode(node.OutgoingConnection.Center)
        // select the X connections here
        // magic - HOW MANY POSSIBLE "TO" NEURONS
        numAxonTerminals := rand.Intn(INITIAL_SYNAPSE_COUNT) + 1
        nodesToConnect := []*Node{
            centralConnNode,
        }
        stDev := DYNAMIC_SYNAPSE_PROB_SPHERE
        for i := 0; i < numAxonTerminals; i++ {
            potPos := [3]int{-1, -1, -1}
            for potPos[0] < 0 || potPos[1] < 0 || potPos[2] < 0 || potPos[0] >= net.Dimensions[0]*2 || potPos[1] >= net.Dimensions[1] || potPos[2] >= net.Dimensions[2] {
                potPos = [3]int{
                    int(rand.NormFloat64() * stDev) + centralConnNode.Position[0],
                    int(rand.NormFloat64() * stDev) + centralConnNode.Position[1],
                    int(rand.NormFloat64() * stDev) + centralConnNode.Position[2],
                }
            }
            potNode := net.FindNode(potPos)
            if !NodeExistsIn(potNode, nodesToConnect) && potNode != node {
                nodesToConnect = append(nodesToConnect, potNode)
            }
        }

        var excitatory bool
        toNodes := make(map[*Node]*ConnInfo)
        for _, node := range nodesToConnect {
            // should this have a higher probability of being excitatory?
            if rand.Intn(INVERSE_INHIBITORY_PROB) != 0 {
                // https://www.quora.com/Are-there-more-excitatory-neurons-or-inhibitory-neurons-in-the-brain-Why
                excitatory = true
            }
            toNodes[node] = &ConnInfo{
                Strength: RandFloat(0.25, 0.75), // magic numbers
                Excitatory: excitatory,
            }
        }

        node.OutgoingConnection.To = toNodes
        for _, nodeToConnect := range nodesToConnect {
            nodeToConnect.IncomingConnections[node] = node.OutgoingConnection
        }
    })
}

func (net *Network) Mirror() {
    // invert in x direction
    leftHemisphere := [][][]*Node{}
    for i := len(net.RightHemisphere)-1; i >= 0; i-- {
        // POINTER CRAPS - NODES
        rightPlane := [][]*Node{}
        for _, rightRow := range net.RightHemisphere[i] {
            aRightRow := []*Node{}
            for _, rightNode := range rightRow {
                newNode := &Node{}
                *newNode = *rightNode
                newNode.IncomingConnections = make(map[*Node]*Connection)
                aRightRow = append(aRightRow, newNode)
            }
            rightPlane = append(rightPlane, aRightRow)
        }
        leftHemisphere = append(leftHemisphere, rightPlane)
    }
    net.LeftHemisphere = leftHemisphere
    net.ForEachRightHemisphereNode(func(node *Node, pos [3]int) {
        actualNode := net.FindLeftHemisphereNode(pos)

        newCenter := node.OutgoingConnection.Center
        newCenter[0] = (net.Dimensions[0]-1) - node.OutgoingConnection.Center[0]

        newCenter[0] += net.Dimensions[0]
        newConn := &Connection{
            Center: newCenter,
        }
        actualNode.OutgoingConnection = newConn
    })
    net.Nodes = append(net.RightHemisphere, net.LeftHemisphere...)
    net.ForEachNode(func(node *Node, pos [3]int) {
        node.Position = pos
        node.Id = fmt.Sprintf("%v|%v|%v", pos[0], pos[1], pos[2])
    })
}

func SumCenterVectors(centers [][3]int, node Node) [3]int {
    final := [3]int{0, 0, 0}
    for _, center := range centers {
        // n - p
        // vector pointing from the center to the node (ie away from center)
        baseVector := [3]float64{float64(node.Position[0] - center[0]), float64(node.Position[1] - center[1]), float64(node.Position[2] - center[2])}
        baseMagnitude := FloatDist(baseVector, [3]float64{0.0, 0.0, 0.0})
        // unit vectorizing
        // d
        d := IntDist(node.Position, center)
        // C = CENTER_RADIUS
        // d - C is distance from node to outer shell
        // d - C > 0 if the node is outside the shell - so make baseVector * negative to point from node to center
        //       < 0 if node is inside shell - baseVector * positive to point from node away from center
        var factor float64
        if CENTER_RADIUS == d {
            factor = 1.0
        } else {
            factor = CENTER_RADIUS/(CENTER_RADIUS - d) * CENTER_VECTOR_FACTOR
        }
        if factor > d {
            factor = d
        }
        if factor < -d {
            factor = -d * CENTER_VECTOR_FACTOR
        }
        for i := 0; i < 3; i++ {
            final[i] += int(baseVector[i]/baseMagnitude * factor)
        }
        if center == node.Position {
            final = [3]int{0, 0, 0}
        }
    }
    return final
}

func (net *Network) Connect() {
    centers := [][3]int{}
    for i := 0; i < NUMBER_OF_CENTERS; i++ {
        centers = append(centers, [3]int{rand.Intn(net.Dimensions[0]), rand.Intn(net.Dimensions[1]), rand.Intn(net.Dimensions[2])})
    }

    net.ForEachRightHemisphereNode(func(node *Node, pos [3]int) {
        // get the closest nodes and select one randomly to connect to
        stDev := AXON_PROB_SPHERE
        center := node.Position
        for center == node.Position {
            potX := int(rand.NormFloat64() * stDev) + node.Position[0]
            potY := int(rand.NormFloat64() * stDev) + node.Position[1]
            potZ := int(rand.NormFloat64() * stDev) + node.Position[2]
            for potX < 0 || potX >= net.Dimensions[0] {
                potX = int(rand.NormFloat64() * stDev) + node.Position[0]
            }
            for potY < 0 || potY >= net.Dimensions[1] {
                potY = int(rand.NormFloat64() * stDev) + node.Position[1]
            }
            for potZ < 0 || potZ >= net.Dimensions[2] {
                potZ = int(rand.NormFloat64() * stDev) + node.Position[2]
            }
            center = [3]int{potX, potY, potZ}
        }

        influenceVector := SumCenterVectors(centers, *node)
        // fmt.Println(influenceVector, node.Position)
        for i := 0; i < 3; i++ {
            center[i] += influenceVector[i]
            if center[i] < 0 {
                center[i] = 0
            }
            if center[i] > net.Dimensions[i] - 1 {
                center[i] = net.Dimensions[i] - 1
            }
        }

        newConn := &Connection{
            Center: center,
        }
        node.OutgoingConnection = newConn
    })
    // fmt.Println(centers)
}

func (net Network) String() string {
    jsonRep, _ := json.MarshalIndent(net, "", "    ")
    return string(jsonRep)
}

func (net Network) DumpJSON(name string, directory string) {
    f, _ := os.Create(fmt.Sprintf("%v/frames/net_%v.json", directory, name))
    f.WriteString(net.String())
    f.Close()
}

func MakeNetwork(dimensions [3]int, blank bool) *Network {
    nodes := [][][]*Node{}
    for i := 0; i < dimensions[0]; i++ {
        iDim := [][]*Node{}
        for j := 0; j < dimensions[1]; j++ {
            jDim := []*Node{}
            for k := 0; k < dimensions[2]; k++ {
                var newValue int
                var randTest float32
                if !blank {
                    randTest = rand.Float32()
                    if randTest < PROB_INITIAL_ON {
                        newValue = 1
                    } else {
                        newValue = 0
                    }
                }
                jDim = append(jDim, &Node{
                    Value: newValue,
                    Position: [3]int{i, j, k},
                    IncomingConnections: make(map[*Node]*Connection),
                    Id: fmt.Sprintf("%v|%v|%v", i, j, k),
                })
            }
            iDim = append(iDim, jDim)
        }
        nodes = append(nodes, iDim)
    }

    return &Network {
        Dimensions: dimensions,
        RightHemisphere: nodes,
        Sensors: make(map[string]*Sensor),
        Outputs: make(map[string]*Output),
        Frames: 0,
    }
}

func (net *Network) GenerateAnim(frames int, directory string) {
    os.Mkdir("frames", 755)
    for frame := 0; frame < frames; frame++ {
        net.DumpJSON(strconv.Itoa(frame), directory)
        net.Cycle()
    }
}