package main

import (
	"fmt"
	"math"
	"os/exec"
	"strings"
	"time"
)

// button struct
type button struct {
	row        int    // topRow or gridRow
	x          int    // collumn index
	y          int    // row index
	color      int    // current button color
	macroColor int    // saved macro led color
	bType      int    // 0: top, 1: right, 2: grid
	pressed    bool   // currently held down
	cmd        string // linux command executed when button gets pressed
}

// button types enum
const (
	TOP = iota
	RIGHT
	GRID
)

// function to turn led at x,y on to specified color
func (b *button) ledOn(color int) error {
	if color > 0 {
		b.color = color
	}
	color = int(math.Abs(float64(color)))
	args := append(pushArgs, fmt.Sprintf("%X %d%d %X", b.row, b.y, b.x, color))
	if b.bType == TOP {
		args = append(pushArgs, fmt.Sprintf("%X %d%X %X", b.row, b.y, b.x+8, color))
	}
	cmd := exec.Command(lpCmd, args...)
	return cmd.Run()

}

// function to turn off led at x,y
func (b *button) ledOff() error {
	b.color = off
	args := append(pushArgs, fmt.Sprintf("%X %d%d 00", b.row, b.y, b.x))
	if b.bType == TOP {
		args = append(pushArgs, fmt.Sprintf("%X %d%X 00", b.row, b.y, b.x+8))
	}
	cmd := exec.Command(lpCmd, args...)
	return cmd.Run()
}

// function to flash a buttons LED n times
func (b *button) flash(color int, n int, delay int) error {
	// repeat n times
	for range n {
		// on
		if err := b.ledOn(-color); err != nil {
			return fmt.Errorf("Error flashing button: %v", err)
		}
		time.Sleep(time.Millisecond * time.Duration(delay))
		// off
		if err := b.ledOn(off); err != nil {
			return fmt.Errorf("Error flashing button: %v", err)
		}
		time.Sleep(time.Millisecond * time.Duration(delay))
	}
	b.ledOn(b.color)
	return nil
}

// function to execute buttons macro command
func (b *button) execute() error {

	// return if button has no command
	if b.cmd == "" {
		return nil
	}

	fmt.Println("EXECUTING COMMAND", b.cmd)

	args := strings.Split(b.cmd, " ")
	// error if no command
	if len(args) == 0 {
		return fmt.Errorf("No command found")
	}

	// set button command
	var cmd *exec.Cmd
	if len(args) == 1 {
		cmd = exec.Command(args[0])
	} else {
		cmd = exec.Command(args[0], args[1:]...)
	}

	// run command
	if err := cmd.Start(); err != nil {
		// flash red and return error
		go b.flash(red, 3, 333)
		return fmt.Errorf("Error starting linux cmd: %v", err)
	}

	// flash green and exit with no error
	b.flash(green, 3, 333/2)
	return nil
}

// function to set the macro for a button
func (b *button) setCMD(command string) error {
	b.cmd = command
	fmt.Printf("Set button %d%d command to %s\n", b.x, b.y, b.cmd)
	return nil
}
