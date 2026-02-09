package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// button struct
type button struct {
	row     string // topRow or gridRow
	x       int    // collumn index
	y       int    // row index
	color   int    // current button color
	bType   int    // 0: top, 1: right, 2: grid
	pressed bool   // currently held down
	cmd     string // linux command executed when button gets pressed
}

// button types enum
const (
	TOP = iota
	RIGHT
	GRID
)

// function to turn led at x,y on to specified color
func (b *button) ledOn(color int) error {
	if color != -1 {
		b.color = color
	}
	args := append(pushArgs, fmt.Sprintf("%s %d%d %d", b.row, b.y, b.x, color))
	if b.bType == TOP {
		args = append(pushArgs, fmt.Sprintf("%s %d%x %d", b.row, b.y, b.x+8, color))
	}
	cmd := exec.Command(lpCmd, args...)
	return cmd.Run()

}

// function to turn off led at x,y
func (b *button) ledOff() error {
	if b.color == off {
		return nil
	}
	b.color = off
	args := append(pushArgs, fmt.Sprintf("%s %d%d 00", b.row, b.y, b.x))
	if b.bType == TOP {
		args = append(pushArgs, fmt.Sprintf("%s %d%x 00", b.row, b.y, b.x+8))
	}
	cmd := exec.Command(lpCmd, args...)
	return cmd.Run()
}

// function to execute buttons macro command
func (b *button) execute() error {
	if b.cmd == "" {
		return nil
	}
	fmt.Println("EXECUTING COMMAND", b.cmd)
	b.ledOn(green)

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
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Error executing button macro: %v", err)
	}
	b.ledOff()
	return nil
}

// function to set the macro for a button
func (b *button) setCMD(command string) error {
	b.cmd = command
	fmt.Printf("Set button %d%d command to %s\n", b.x, b.y, b.cmd)
	return nil
}
