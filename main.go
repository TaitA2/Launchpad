package main

import (
	"errors"
	"fmt"
	"log"
	"os"
)

// set path for the csv file containing macros
var macroDir = ".config/launchpad/"
var macroFile = "commands.csv"

// set 'amidi' as the linux command to use for communicating with the launchpad
var lpCmd string = "amidi"
var getArgs []string
var pushArgs []string

// LED color codes
const off = 0
const red = 3
const green = 48
const amber = 35
const lime = 50

// default user color
const defaultColor = amber

// map of string color names to color codes
var colors = map[string]int{"green": green, "red": red, "amber": amber, "lime": lime}

func main() {
	// setup config
	if err := setConfig(); err != nil {
		log.Fatalf("Error setting up config: %v", err)
	}
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

// function to find / create macro command config file
func setConfig() error {
	// set config path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("Error finding user home dir: %v", err)
	}
	macroFile = homeDir + "/" + macroDir + macroFile

	// create the file
	if _, err := os.Stat(macroFile); errors.Is(err, os.ErrNotExist) {
		if err := os.Mkdir(homeDir+"/"+macroDir, 0777); err != nil {
			return fmt.Errorf("Error creating config directory: %v", err)
		}
		f, err := os.Create(macroFile)
		if err != nil {
			return fmt.Errorf("Error creating config file: %v", err)
		}
		f.Close()
	}

	// change to home dir
	return os.Chdir(homeDir)
}
