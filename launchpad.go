package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
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
	topButtons   []*button      // array x index to topRow buttons
	rightButtons []*button      // array y index of right collumn buttons
	gridButtons  [][]*button    // 2D array of buttons - first index for row, second index for collumn
	buttonChan   chan *button   // channel for current button
	layerCMDs    []func() error // array of layer functions
	layer        int            // current active 'layer' (0-7) tied to top row
	userColor    int            // current color selected by user
}

// function to start the launchpad
func (lp *launchpad) start() error {

	fmt.Println("Started launchpad!")

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
	for {
		// run current layers command
		if err := lp.layerCMDs[lp.layer](); err != nil {
			return err
		}
		// layer has changed
		if prevLayer != lp.layer {
			fmt.Printf("Switching to layer: %d!\n", lp.layer)
			// clear grid unless layer is freeze or paint
			if lp.layer != FREEZE && lp.layer != PAINT {
				lp.gridOff()
			}
			if lp.layer == RECORD {
				go lp.macroFlash()
			}
			// update previous layer var
			prevLayer = lp.layer
		}
		// enable led of current layer
		lp.topButtons[lp.layer].ledOn(lp.userColor)
	}
}

// function to turn on led of any buttons with a set command
func (lp *launchpad) macroLights() error {
	for _, row := range lp.gridButtons {
		for _, b := range row {
			if b.cmd != "" {
				b.ledOn(lp.userColor)
			}
		}
	}

	return nil
}

// function to flash grid buttons with macro command
func (lp *launchpad) macroFlash() {
	for lp.layer == RECORD {
		lp.gridOff()
		time.Sleep(time.Millisecond * 500)
		lp.macroLights()
		time.Sleep(time.Millisecond * 500)
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

	// initialise button channel
	lp.buttonChan = make(chan *button, 160)

	// initialise button arrays
	fmt.Println("Creating buttons...")
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
	fmt.Println("Setting up layers...")
	lp.setLayerCMDs()

	// get macros
	fmt.Println("Setting up macros...")
	if err := lp.getMacros(); err != nil {
		return nil, err
	}

	// return a pointer to the launchpad
	return &lp, nil
}

// function to load macros
func (lp *launchpad) getMacros() error {
	file, err := os.Open(macroFile)
	if err != nil {
		return fmt.Errorf("Error opening macro file: %v", err)
	}
	defer file.Close()

	// create new scanner
	scanner := bufio.NewScanner(file)

	// header row
	scanner.Scan()

	// command rows
	for scanner.Scan() {
		// get row, column, and command string
		info := strings.Split(scanner.Text(), ",")
		if len(info) != 3 {
			return fmt.Errorf("Error scanning commands file, invalid line: %v", info)
		}
		row, err := strconv.Atoi(info[0])
		if err != nil {
			return fmt.Errorf("Error converting %s to a row: %v", info[0], err)
		}
		col, err := strconv.Atoi(info[1])
		if err != nil {
			return fmt.Errorf("Error converting %s to a column: %v", info[0], err)
		}

		cmd := info[2]

		// set button command
		lp.gridButtons[row][col].cmd = cmd
		fmt.Printf("Set button at row: %d, col: %d to '%s'.\n", row, col, cmd)

	}

	// exit without error
	return nil
}

// function to save macros to a file
func (lp *launchpad) saveMacros() error {
	file, err := os.Create(macroFile)
	if err != nil {
		return fmt.Errorf("Error opening macro file: %v", err)
	}
	defer file.Close()
	if _, err := fmt.Fprint(file, "row,column,cmd\n"); err != nil {
		return fmt.Errorf("Error writing macro file header: %v", err)
	}

	// iterate over all grid buttons
	for i, row := range lp.gridButtons {
		for j, b := range row {
			if b.cmd != "" {
				if _, err := fmt.Fprintf(file, "%d,%d,%s\n", i, j, b.cmd); err != nil {
					return fmt.Errorf("Error writing to macro file: %v", err)
				}
			}
		}
	}

	file.Close()
	fmt.Printf("Saved macros to file: %s!\n", file.Name())
	// exit without error
	return nil
}

// function to set the midi path used by the launchpad
func getMidi() error {
	fmt.Println("Finding midi port for the launchpad.")

	// list midi devices
	cmd := exec.Command(lpCmd, "-l")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Error getting midi path to launchpad: %v", err)
	}

	// split output into seperate lines
	lines := strings.Split(string(out), "\n")
	if len(lines) < 2 {
		// error if no devices foudn
		return fmt.Errorf("Could not find launchpad in midi devices: %v", lines)
	}

	// iterate through output
	for _, line := range lines {
		// iterate through line containing 'Launchpad'
		if strings.Contains(line, "Launchpad") {
			for s := range strings.SplitSeq(line, " ") {

				// find path in format "hw:x,x,x"
				if strings.Contains(s, "hw") {
					// set global variables
					path := s
					getArgs = []string{"-p", path, "-d"}
					pushArgs = []string{"-p", path, "-S"}
					fmt.Println("Found path for launchpad as: ", path)

					// exit without error
					return nil
				}
			}
		}
	}

	// error if not found
	return fmt.Errorf("Could not find midi path for launchpad")
}

// function to set top row layers
func (lp *launchpad) setLayerCMDs() {
	lp.layerCMDs = make([]func() error, 8)

	lp.layerCMDs[FREEZE] = lp.freeze
	lp.layerCMDs[PAINT] = lp.paint
	lp.layerCMDs[BREATHE] = lp.breathe
	lp.layerCMDs[ALL] = lp.gridOn
	lp.layerCMDs[MACRO] = lp.macro
	lp.layerCMDs[RECORD] = lp.recordMacro

	// UNIMPLEMENTED
	lp.layerCMDs[6] = lp.colorDebug
	lp.layerCMDs[7] = lp.freeze
}

// function to freeze launchpad LEDs as they are
func (lp *launchpad) freeze() error {

	// get button
	b := lp.getBtn()
	// exit if not a grid button
	if b.bType != GRID {
		return nil
	}

	// reset button when released
	if !b.pressed {
		return b.ledOn(b.color)

	}

	// change the LED to a different color. use negative value to not save color
	if b.color != lp.userColor {
		return b.ledOn(-lp.userColor) // enable LED to user color
	}

	// always use different color than current
	if b.color == lime {
		return b.ledOn(-amber)
	}
	return b.ledOn(-lime)

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
				// turn off led for old layer
				lp.topButtons[lp.layer].ledOff()
				// switch layer
				lp.layer = b.x
				// turn on led for new layer
				lp.topButtons[lp.layer].ledOn(lp.userColor)
			}
			// change color for right button
		} else if y == 8 {
			b = lp.rightButtons[x]
			lp.userColor = b.color
			// fmt.Println("Switching color to", lp.userColor)

		} else {
			b = lp.gridButtons[x][y]
		}
		b.pressed = pressed
		lp.buttonChan <- b
	}
}

// function to get one line of launchpad input
func (lp *launchpad) getBtn() *button {
	return <-lp.buttonChan
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
		if lp.userColor == off {
			return b.ledOff()
		}
		b.ledOn(lp.userColor)
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
	lp.rightButtons[0].ledOn(0)
	lp.rightButtons[1].ledOn(48)
	lp.rightButtons[2].ledOn(49)
	lp.rightButtons[3].ledOn(50)
	lp.rightButtons[4].ledOn(51)
	lp.rightButtons[5].ledOn(35)
	lp.rightButtons[6].ledOn(19)
	lp.rightButtons[7].ledOn(3)

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

// function to display all possible colors
func (lp *launchpad) colorDebug() error {
	// fill grid with colors
	k := 0
	for i := range 4 {
		for j := range 4 {
			lp.gridButtons[i][j].ledOn(k)
			k++
		}
		k += 12
	}
	// print pressed buttons color
	b := lp.getBtn()
	fmt.Printf("Dec: %d, Hex: %x\n", b.color, b.color)
	return nil
}

// layer to execute linux cmd of button pushed
func (lp *launchpad) macro() error {
	lp.macroLights()
	// get current button
	b := lp.getBtn()
	if b.bType != GRID {
		return nil
	}

	// if button has no macro
	if b.cmd == "" {
		// enable LED when pressed
		if b.pressed {
			b.ledOn(lp.userColor)
			return nil
		}
		// disable LED when released
		b.ledOff()
		return nil
	}

	// button has a macro

	// run the macro when pressed
	if b.pressed {
		if err := b.execute(); err != nil {
			log.Printf("Error executing macro: %v", err)
		}
	}

	// reset macro button color
	b.ledOn(lp.userColor)

	// exit without error
	return nil
}

// layer to set the macro of button pushed
func (lp *launchpad) recordMacro() error {
	// get current button
	b := lp.getBtn()
	if b.bType != GRID || !b.pressed {
		return nil
	}

	// light LED
	go b.flash(lp.userColor, 3, 200)

	// scan for input
	c, err := getInputFromPopup()
	if err != nil {
		go b.flash(red, 3, 333/2)
		return fmt.Errorf("Error getting user input: %v", err)
	}
	// save command to button
	b.cmd = c

	// give approval
	fmt.Println("Command set to: ", b.cmd)
	defer b.flash(green, 3, 333/2)

	// save new macros to file and return any error
	return lp.saveMacros()
}

// Function to get input from terminal
func getInputFromPopup() (string, error) {
	// Create a temporary file to store user input
	tempFile, err := os.CreateTemp("", "input.txt")
	if err != nil {
		return "", err
	}
	defer os.Remove(tempFile.Name()) // Clean up the temp file afterwards

	// Launch the kitty terminal command to read user input
	cmd := exec.Command("kitty", "--title", "Launchpad Input", "--", "bash", "-c", fmt.Sprintf("read -p 'Enter command to save: ' userInput; echo $userInput > '%s'", tempFile.Name()))

	if err := cmd.Run(); err != nil {
		return "", err
	}

	// Read the user input from the temporary file
	o, err := os.ReadFile(tempFile.Name())
	if err != nil {
		return "", err
	}
	out := string(o)

	return strings.Trim(out, "\n"), nil
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
