package main

var lpCmd string = "amidi"
var getArgs = []string{"-d", "-p", "hw:0,0,0"}
var pushArgs = []string{"-p", "hw:0,0,0", "-S"}

// LED color codes
const off = 0
const green = 30
const red = 3
const amber = 13
const lime = 39

// map of string color names to color codes
var colors = map[string]int{"green": green, "red": red, "amber": amber, "lime": lime}

// var layerCMDs  []func() = {test, all_on, flash, paint}

func main() {
	// get the launchpad struct
	lp := get_launchpad()

	// start launchpad
	lp.start()

}
