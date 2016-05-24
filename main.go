package main

import (
    "os"
    "fmt"
    "time"
    "bufio"
    "strings"
    // "reflect"
    "math/rand"

    "github.com/jteeuwen/keyboard/termbox"
    term "github.com/nsf/termbox-go"
)

/*
    & is "address of"
    * is "value at address"
*/

var running = true

func main() {
    reader := bufio.NewReader(os.Stdin)
    fileName := Prompt("Enter state name to load state, or leave blank to create a new network:  ", reader)

    rand.Seed(time.Now().UTC().UnixNano())

    var myNet *Network
    _, err := os.Stat(fmt.Sprintf("./state/%v_state.json", fileName))
    if fileName == "" || err != nil {
        NETWORK_SIZE := [3]int{25, 25, 25}
        myNet = MakeNetwork(NETWORK_SIZE, false)
        myNet.Connect()
        // myNet.CreateSensor("a", 1, 50, "", [3]int{25, 1, 1}, true, "a", kb)
        // myNet.CreateSensor("s", 1, 50, "", [3]int{1, 1, 1}, true, "s", kb)
        // myNet.CreateSensor("d", 1, 50, "", [3]int{12, 12, 1}, true, "d", kb)
        // myNet.CreateSensor("f", 1, 50, "", [3]int{12, 1, 12}, true, "f", kb)
        // myNet.CreateOutput("12, 1, 1", 1, 50,"", [3]int{12, 1, 1})
    } else {
        myNet = LoadState(fileName)
    }

    var choice string
    fmt.Printf("Network has %v sensor(s):\n", len(myNet.Sensors))
    for _, sensor := range myNet.Sensors {
        fmt.Printf("    %v\n", sensor.Name)
    }
    choice = Prompt("\nAdd sensor? [y/n]  ", reader)
    for choice == "y" {
        sensorName := Prompt("    Name:  ", reader)

        trigger := Prompt("    Trigger [single key]:  ", reader) // should validate to be one key

        plane := Prompt("    Plane [x/y/z/blank]:  ", reader)
        if plane != "x" && plane != "y" && plane != "z" {
            plane = ""
        }

        // validate for negatives
        centerArr := []int{}
        for len(centerArr) != 3 {
            center := Prompt("    Center [format x,y,z]:  ", reader)
            centerArr = StrsToInts(strings.Split(center, ","))
        }

        myNet.CreateSensor(sensorName, 1, 50, plane, [3]int{centerArr[0], centerArr[1], centerArr[2]}, true, trigger) // todo find numbers and stuff

        choice = Prompt("Add another sensor? [y/n]  ", reader)
    }

    choice = Prompt("\nEnter a sensor name to remove a sensor:  ", reader)
    for choice != "" {
        myNet.RemoveSensor(choice)
        choice = Prompt("Enter another sensor name to remove:  ", reader)
    }

    fmt.Printf("Network has %v output(s).\n", len(myNet.Outputs))
    for _, output := range myNet.Outputs {
        fmt.Printf("    %v\n", output.Name)
    }
    choice = Prompt("\n    Add output? [y/n]  ", reader)
    for choice == "y" {
        outputName := Prompt("    Name:  ", reader)

        plane := Prompt("    Plane [x/y/z/blank]:  ", reader)
        if plane != "x" && plane != "y" && plane != "z" {
            plane = ""
        }

        // validate for negatives
        centerArr := []int{}
        for len(centerArr) != 3 {
            center := Prompt("    Center [format x,y,z]:  ", reader)
            centerArr = StrsToInts(strings.Split(center, ","))
        }

        myNet.CreateOutput(outputName, 1, 50, plane, [3]int{centerArr[0], centerArr[1], centerArr[2]}) //todo get numbers

        choice = Prompt("Add another output? [y/n]  ", reader)
    }
    choice = Prompt("\n    Enter an output name to remove an output:  ", reader)
    for choice != "" {
        myNet.RemoveOutput(choice)
        choice = Prompt("Enter another output name to remove:  ", reader)
    }

    // this is the keyboard sensing stuff
    term.Init()
    term.SetCursor(0, 0)
    
    kb := termbox.New()
    kb.Bind(func() {
        running = false
    }, "space")
    go KeyboardPoll(kb)
    myNet.BindKeyboard(kb)

    myNet.AnimateUntilDone(100)

    term.Close()

    fmt.Print("\nSave state?  Enter a name if you wish to save the state:  ")
    fileName, _ = reader.ReadString('\n')
    fileName = strings.TrimSpace(fileName)
    if fileName != "" {
        myNet.SaveState(fileName)
    }

    // this section is to test state saving/loading capabilities
    // NETWORK_SIZE := [3]int{25, 25, 25}
    // myNet := MakeNetwork(NETWORK_SIZE, false)
    // myNet.Connect()

    // myNet.CreateSensor("aa", 1, 50, "", [3]int{24, 0, 0}, true, "a", nil)
    // myNet.CreateSensor("bb", 1, 50, "", [3]int{0, 0, 0}, true, "b", nil)
    // myNet.CreateOutput("output", 1, 50,"", [3]int{12, 1, 1})
    // myNet.SaveState("test")
    // loadedNet := LoadState("test", nil)
    // fmt.Println(reflect.DeepEqual(loadedNet, myNet))
}