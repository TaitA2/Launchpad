package main

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// top row vs grid row codes
const topRow = "B0"
const gridRow = "90"

// layer command enum
const (
	TEST = iota
	ALL
	FLASH
	PAINT
)

// launchpad struct
type launchpad struct {
	topButtons    []*button      // array x index to topRow buttons
	rightButtons  []*button      // array y index of right collumn buttons
	gridButtons   [][]*button    // 2D array of buttons - first index for row, second index for collumn
	pressedButton *button        // array of pressed buttons
	layerCMDs     []func() error // array of layer functions
	layer         int            // current active 'layer' (0-7) tied to top row
	userColor     int            // current color selected by user
}

// function to return launchpad struct
func get_launchpad() *launchpad {

	// initialise launchpad
	var lp launchpad

	// initialise button arrays
	lp.topButtons = make([]*button, 8)
	lp.rightButtons = make([]*button, 8)
	lp.gridButtons = make([][]*button, 8)

	// populate button arrays with coords and default values
	for i := range 8 {
		lp.gridButtons[i] = make([]*button, 8)
		topBtn := button{row: topRow, x: i, y: 6, color: off, pressed: false, bType: TOP}
		lp.topButtons[i] = &topBtn
		rightBtn := button{row: gridRow, x: 8, y: i, color: off, pressed: false, bType: RIGHT}
		lp.rightButtons[i] = &rightBtn
		for j := range 8 {
			gridBtn := button{row: gridRow, x: j, y: i, color: off, pressed: false, bType: GRID}
			lp.gridButtons[i][j] = &gridBtn
		}
	}

	// get layer functions
	lp.getLayerCMDs()

	// return a pointer to the launchpad
	return &lp
}

func (lp *launchpad) getLayerCMDs() {
	lp.layerCMDs = make([]func() error, 8)
	lp.layerCMDs[0] = lp.push_test
	lp.layerCMDs[1] = lp.grid_on
	lp.layerCMDs[2] = lp.flash_grid
	lp.layerCMDs[3] = lp.paint
}

func (lp *launchpad) start() error {
	fmt.Println("Starting launchpad!")
	lp.allOff()
	lp.pallette()
	go lp.listen()
	prevLayer := -1
	lp.topButtons[lp.layer].ledOn(lp.userColor)
	for {
		if prevLayer != lp.layer {
			fmt.Printf("Switching to layer: %d!\n", lp.layer)
			prevLayer = lp.layer
		}
		// run current layers command
		if err := lp.layerCMDs[lp.layer](); err != nil {
			return err
		}
	}
}

// function to flash all leds
func (lp *launchpad) flash_grid() error {
	lp.grid_on()
	time.Sleep(time.Second)
	lp.grid_off()
	time.Sleep(time.Second)
	return nil
}

// function to constantly monitor launchapd input
func (lp *launchpad) listen() error {
	// loop forever
	for {
		// create midi command
		cmd := exec.Command(lpCmd, getArgs...)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Fatalf("Error creating stdout: %v", err)
		}
		// start the midi command
		cmd.Start()

		// create buffer for the output
		var output = make([]byte, 9)
		n := 0
		for n < 9 {
			n, _ = stdout.Read(output)
		}

		// kill the midi command process
		cmd.Process.Kill()

		// break output into parts
		parts := strings.Split(string(output), " ")

		// create button struct from midi
		var b *button
		row := parts[0]
		x, _ := strconv.Atoi(parts[1][0:1])
		y, _ := strconv.ParseInt(parts[1][1:], 16, 64)
		pressed := parts[2] != "00"

		// change layer for top button
		if strings.Contains(row, topRow) {
			b = lp.topButtons[y-8]
			if pressed && b.x < len(lp.layerCMDs) {
				lp.grid_off()
				lp.topButtons[lp.layer].ledOff()
				lp.layer = b.x
				lp.topButtons[lp.layer].ledOn(lp.userColor)
			}
			// change color for right button
		} else if y == 8 {
			b = lp.rightButtons[x]
			if pressed {
				lp.userColor = b.color
				// fmt.Println("Switching color to", lp.userColor)
			}
		} else {
			b = lp.gridButtons[x][y]
		}
		b.pressed = pressed

		lp.pressedButton = b
	}
}

// function to get one line of launchpad input
func (lp *launchpad) getBtn() *button {
	for lp.pressedButton == nil {
	}
	return lp.pressedButton
}

// turn off all grid buttons
func (lp *launchpad) grid_off() error {
	for _, row := range lp.gridButtons {
		for _, btn := range row {
			btn.color = off
			if err := btn.ledOff(); err != nil {
				return err
			}
		}
	}
	return nil
}

// turn on all grid buttons
func (lp *launchpad) grid_on() error {
	for _, row := range lp.gridButtons {
		for _, btn := range row {
			if err := btn.ledOn(lp.userColor); err != nil {
				return err
			}
		}
	}
	return nil
}

// turn on all right collumn buttons
func (lp *launchpad) right_on() error {
	for _, b := range lp.rightButtons {
		if err := b.ledOn(b.color); err != nil {
			return err
		}
	}
	return nil
}

// turn off all right collumn buttons
func (lp *launchpad) right_off() error {
	for _, b := range lp.rightButtons {
		if err := b.ledOff(); err != nil {
			return err
		}
	}
	return nil
}

// function to turn all leds on to specified color
func (lp *launchpad) all_on() error {

	color := lp.userColor
	// turn on all top buttons
	for _, btn := range lp.topButtons {
		if err := btn.ledOn(color); err != nil {
			return err
		}
	}

	if err := lp.right_on(); err != nil {
		return err
	}

	if err := lp.grid_on(); err != nil {
		return err
	}
	return nil
}

// function to turn all leds on to specified color
func (lp *launchpad) allOff() error {
	// turn off all top buttons
	for i := range lp.topButtons {
		if err := lp.topButtons[i].ledOff(); err != nil {
			return err
		}
		if err := lp.rightButtons[i].ledOff(); err != nil {
			return err
		}
	}

	// turn off all grid buttons
	for i, row := range lp.gridButtons {
		for j := range row {
			if err := lp.gridButtons[i][j].ledOff(); err != nil {
				return err
			}
		}
	}
	return nil
}

// function to turn on any pushed leds
func (lp *launchpad) paint() error {

	// get the pressed button
	b := lp.getBtn()

	// add color for grid button
	if b.bType == GRID {
		if err := b.ledOn(lp.userColor); err != nil {
			return err
		}
	}

	// exit without error
	return nil
}

// function to setup the paint environment
func (lp *launchpad) pallette() {
	// turn off all LEDs
	lp.grid_off()
	lp.userColor = green

	// set right buttons as color pallette
	lp.rightButtons[0].ledOn(off)
	lp.rightButtons[1].ledOn(green)
	lp.rightButtons[2].ledOn(lime)
	lp.rightButtons[3].ledOn(amber)
	lp.rightButtons[4].ledOn(red)

}

// function to enable LED of any button while its pushed
func (lp *launchpad) push_test() error {

	b := lp.getBtn()

	if b.bType != GRID {
		return nil
	}

	if b.pressed {
		b.ledOn(lp.userColor)
	} else {
		b.ledOff()
	}
	return nil
}

func (lp *launchpad) colorDebug() {
	fmt.Println("Filling colors...")
	k := 0
	for {
		start := k
		for i := range 8 {
			for j := range 4 {
				lp.gridButtons[i][j].ledOn(k)
				k++
			}
		}
		fmt.Printf("Showing %d - %d\n", start, k)
		for i := range 8 {
			for j := range 4 {
				lp.gridButtons[i][j+4].ledOn(k)
				k++
			}
		}
		time.Sleep(time.Second * 5)
	}
}
