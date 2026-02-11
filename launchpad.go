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
	FREEZE = iota
	PAINT
	BREATHE
	ALL
	MACRO
	RECORD
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

// function to start the launchpad
func (lp *launchpad) start() error {

	fmt.Println("Starting launchpad!")

	// draw startup flower spash
	lp.drawFlower()
	time.Sleep(time.Second * 1)

	// clear LEDs and enable color selector pallette
	lp.allOff()
	lp.pallette()

	// start listening for button events
	go lp.listen()
	prevLayer := 0
	lp.topButtons[lp.layer].ledOn(lp.userColor)
	lp.gridButtons[0][0].setCMD("kitty -e sl")
	lp.gridButtons[7][0].setCMD("firefox tidal.com")
	lp.gridButtons[7][1].setCMD("firefox airtable.com")
	for {
		// run current layers command
		if err := lp.layerCMDs[lp.layer](); err != nil {
			return err
		}
		if prevLayer != lp.layer {
			fmt.Printf("Switching to layer: %d!\n", lp.layer)
			if !((lp.layer == FREEZE) || (lp.layer == PAINT && prevLayer == FREEZE)) {
				lp.gridOff()
			}
			// lp.topButtons[prevLayer].ledOff()
			prevLayer = lp.layer
		}
		lp.topButtons[lp.layer].ledOn(lp.userColor)
	}
}

// function to return launchpad struct
func getLaunchpad() (*launchpad, error) {

	// initialise launchpad
	var lp launchpad

	// get path to midi device
	if err := getMidi(); err != nil {
		return nil, err
	}

	// initialise button arrays
	lp.topButtons = make([]*button, 8)
	lp.rightButtons = make([]*button, 8)
	lp.gridButtons = make([][]*button, 8)

	// populate button arrays with coords and default values
	for i := range 8 {
		lp.gridButtons[i] = make([]*button, 8)
		lp.topButtons[i] = &button{row: topRow, x: i, y: 6, color: green, pressed: false, bType: TOP}
		lp.rightButtons[i] = &button{row: gridRow, x: 8, y: i, color: green, pressed: false, bType: RIGHT}
		for j := range 8 {
			lp.gridButtons[i][j] = &button{row: gridRow, x: j, y: i, color: green, pressed: false, bType: GRID}
		}
	}

	// get layer functions
	lp.getLayerCMDs()

	// return a pointer to the launchpad
	return &lp, nil
}

// function to set the midi path used by the launchpad
func getMidi() error {
	cmd := exec.Command(lpCmd, "-l")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Error getting midi path to launchpad: %v", err)
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) < 2 {
		return fmt.Errorf("Could not find launchpad in midi devices: %v", lines)
	}
	for _, line := range lines {
		if strings.Contains(line, "Launchpad") {
			path := strings.Split(line, " ")[2]
			fmt.Println("Found path for launchpad as: ", path)
			getArgs = []string{"-p", path, "-d"}
			pushArgs = []string{"-p", path, "-S"}
			return nil
		}
	}

	return fmt.Errorf("Could not find midi path for launchpad")
}

func (lp *launchpad) getLayerCMDs() {
	lp.layerCMDs = make([]func() error, 8)
	lp.layerCMDs[FREEZE] = lp.freeze
	lp.layerCMDs[PAINT] = lp.paint
	lp.layerCMDs[BREATHE] = lp.breathe
	lp.layerCMDs[ALL] = lp.gridOn
	lp.layerCMDs[MACRO] = lp.macro
}

// function to freeze launchpad LEDs as they are
func (lp *launchpad) freeze() error {
	b := lp.getBtn()
	color := b.color
	for b != nil && b.pressed {
		if b.color == lp.userColor {
			if b.color == lime {
				b.ledOn(amber)
			} else {
				b.ledOn(lime)
			}

		} else {
			b.ledOn(lp.userColor)
		}
		oldB := b
		b = lp.getBtn()
		if oldB != b {
			oldB.ledOn(color)
			color = b.color
		}
	}
	b.ledOn(color)

	return nil
}

// function to flash all leds
func (lp *launchpad) strobe() error {
	time.Sleep(time.Millisecond * 100)
	lp.gridOff()
	time.Sleep(time.Millisecond * 100)
	return lp.gridOn()

}

// function to ex/implode all leds
func (lp *launchpad) breathe() error {
	lp.implodeOff()
	time.Sleep(time.Millisecond * 500)
	lp.explodeOn()
	time.Sleep(time.Millisecond * 500)
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
				// refresh grid when same layer pressed
				if b.x == lp.layer {
					lp.gridOff()
				}
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
	b := lp.pressedButton
	lp.pressedButton = nil
	return b
}

// turn off all grid buttons
func (lp *launchpad) gridOff() error {
	for _, row := range lp.gridButtons {
		for _, btn := range row {
			if err := btn.ledOff(); err != nil {
				return err
			}
		}
	}
	return nil
}

// function to turn off all grid LEDs outside in
func (lp *launchpad) implodeOff() error {
	for i := range len(lp.gridButtons) / 2 {
		for j := range len(lp.gridButtons[i]) / 2 {

			btn := lp.gridButtons[j][i]
			if err := btn.ledOff(); err != nil {
				return err
			}
			btn = lp.gridButtons[j][7-i]
			if err := btn.ledOff(); err != nil {
				return err
			}
			btn = lp.gridButtons[7-j][i]
			if err := btn.ledOff(); err != nil {
				return err
			}
			btn = lp.gridButtons[7-j][7-i]
			if err := btn.ledOff(); err != nil {
				return err
			}
		}
	}

	return nil
}

// function to turn off all grid LEDs outside in
func (lp *launchpad) implodeOn() error {
	for i := range len(lp.gridButtons) / 2 {
		for j := range len(lp.gridButtons[i]) / 2 {

			btn := lp.gridButtons[j][i]
			if err := btn.ledOn(lp.userColor); err != nil {
				return err
			}
			btn = lp.gridButtons[j][7-i]
			if err := btn.ledOn(lp.userColor); err != nil {
				return err
			}
			btn = lp.gridButtons[7-j][i]
			if err := btn.ledOn(lp.userColor); err != nil {
				return err
			}
			btn = lp.gridButtons[7-j][7-i]
			if err := btn.ledOn(lp.userColor); err != nil {
				return err
			}
		}
	}

	return nil
}

// function to turn off all grid LEDs outside in
func (lp *launchpad) explodeOn() error {
	for i := range len(lp.gridButtons) / 2 {
		for j := range len(lp.gridButtons[i]) / 2 {

			btn := lp.gridButtons[3-j][3-i]
			if err := btn.ledOn(lp.userColor); err != nil {
				return err
			}
			btn = lp.gridButtons[3-j][4+i]
			if err := btn.ledOn(lp.userColor); err != nil {
				return err
			}
			btn = lp.gridButtons[4+j][3-i]
			if err := btn.ledOn(lp.userColor); err != nil {
				return err
			}
			btn = lp.gridButtons[4+j][4+i]
			if err := btn.ledOn(lp.userColor); err != nil {
				return err
			}
		}
	}

	return nil
}

// turn on all grid buttons
func (lp *launchpad) gridOn() error {
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
func (lp *launchpad) rightOn() error {
	for _, b := range lp.rightButtons {
		if err := b.ledOn(b.color); err != nil {
			return err
		}
	}
	return nil
}

// turn off all right collumn buttons
func (lp *launchpad) rightOff() error {
	for _, b := range lp.rightButtons {
		if err := b.ledOff(); err != nil {
			return err
		}
	}
	return nil
}

// function to turn all leds on to specified color
func (lp *launchpad) forceAllOn() error {
	args := append(pushArgs, fmt.Sprintf("B0 00 %d", lp.userColor))
	cmd := exec.Command(lpCmd, args...)
	return cmd.Run()
}

// function to turn all leds on to specified color
func (lp *launchpad) allOn() error {

	// turn on all top buttons
	for _, btn := range lp.topButtons {
		if err := btn.ledOn(lp.userColor); err != nil {
			return err
		}
	}

	if err := lp.rightOn(); err != nil {
		return err
	}

	if err := lp.gridOn(); err != nil {
		return err
	}
	return nil
}

// function to turn all leds on to specified color
func (lp *launchpad) forceAllOff() error {
	args := append(pushArgs, "B0 00 00")
	cmd := exec.Command(lpCmd, args...)
	return cmd.Run()
}

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
	lp.gridOff()
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
	lp.gridOff()

	// initialise the current color as the default color
	lp.userColor = defaultColor

	// set right buttons as color pallette
	lp.rightButtons[0].ledOn(off)
	lp.rightButtons[1].ledOn(green)
	lp.rightButtons[2].ledOn(lime)
	lp.rightButtons[3].ledOn(amber)
	lp.rightButtons[4].ledOn(red)

}

// function to enable LED of any button while its pushed
func (lp *launchpad) pushTest() error {

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

// function to execute linux cmd of button pushed
func (lp *launchpad) macro() error {
	// get current button
	b := lp.getBtn()
	if b.pressed {
		if b.cmd == "" {
			b.ledOn(lp.userColor)
		} else {
			return b.execute()
		}
	} else if b.bType == GRID {
		b.ledOff()
	} else {
		b.ledOn(lp.userColor)
	}
	return nil
}

func (lp *launchpad) drawFlower() error {
	lp.allOff()
	// top, bottom and middle of flower
	for k := range 2 {
		// lime core
		lp.gridButtons[2][3+k].ledOn(lime)
		// amber sides
		lp.gridButtons[2][2+(3*k)].ledOn(amber)

		// green leaves
		lp.gridButtons[5][k].ledOn(green)
		lp.gridButtons[5][k+6].ledOn(green)
		lp.gridButtons[6][2+(k*3)].ledOn(green)

		// top and bottom
		for i := range 4 {
			lp.gridButtons[k*4][i+2].ledOn(red)
			lp.gridButtons[k*4+(1+(k*-2))][i+2].ledOn(amber)

		}
	}

	for k := range 3 {
		// red walls
		lp.gridButtons[k+1][1].ledOn(red)
		lp.gridButtons[k+1][6].ledOn(red)

		// green stem
		lp.gridButtons[k+5][3].ledOn(green)
		lp.gridButtons[k+5][4].ledOn(green)
	}
	return nil
}

func (lp *launchpad) flashFlower() error {
	lp.drawFlower()
	time.Sleep(time.Millisecond * 200)
	lp.forceAllOff()
	time.Sleep(time.Millisecond * 200)

	return nil
}
