potentially - save sensor/output centers (and possibly nodes?) but not the functions, and have user input functions?  that would be after this branch is merged though.

go through all of my `// todo - ` comments and clean up anything unnecessary

https://www.raspberrypi.org/forums/viewtopic.php?f=29&t=109587

### working on

exploring the possibility of saving centers(and nodes?  speed diff?) of outputs/sensors so you can reset the functions

right now, output functions are just additive, so that wouldn't be a big deal

the output connections *are* unique to the output 

would have to check the same node coordinates in save_test.go; however, would not check whether function esists

also display existing sensors (not outputs) to user on load so that user can cancel/restart based on that information

after custom functions are loaded, runs cript to prune sensors (and their corresponding outputs) that don't have their functions set