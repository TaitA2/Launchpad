package main

import (
	"fmt"
	"os/exec"
)

// button struct
type button struct {
	row     string   // topRow or gridRow
	x       int      // collumn index
	y       int      // row index
	color   int      // current button color
	bType   int      // 0: top, 1: right, 2: grid
	pressed bool     // currently held down
	cmd     exec.Cmd // linux command executed when button gets pressed
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
	args := append(pushArgs, fmt.Sprintf("%s %d%d 00", b.row, b.y, b.x))
	if b.bType == TOP {
		args = append(pushArgs, fmt.Sprintf("%s %d%x 00", b.row, b.y, b.x+8))
	}
	cmd := exec.Command(lpCmd, args...)
	return cmd.Run()
}
