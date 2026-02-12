package main

import (
	"fmt"
	"log"
	"os"
)

// set path for the csv file containing macros
var macroFile = ".config/launchpad/commands.csv"

// set 'amidi' as the linux command to use for communicating with the launchpad
var lpCmd string = "amidi"
var getArgs []string
var pushArgs []string

// LED color codes
const off = 0
const green = 30
const red = 3
const amber = 13
const lime = 39

// default user color
const defaultColor = amber

// map of string color names to color codes
var colors = map[string]int{"green": green, "red": red, "amber": amber, "lime": lime}

func main() {
	// set config path
	dir, err := os.UserHomeDir()
	macroFile = dir + "/" + macroFile
	// change to home dir
	os.Chdir(dir)

	// get the launchpad struct
	fmt.Println("Getting launchpad...")
	lp, err := getLaunchpad()
	if err != nil {
		log.Fatalf("Error getting launchpad: %v", err)
	}

	// start launchpad
	if err := lp.start(); err != nil {
		log.Fatalf("Error starting launchpad: %v", err)
	}

}
