package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

var lpCmd string = "amidi"
var getArgs = []string{"-d", "-p", "hw:0,0,0"}
var pushArgs = []string{"-p", "hw:0,0,0", "-S"}

const green = "30"
const red = "3"
const amber = "13"
const lime = "39"

func main() {
	all_off()
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("Too few arguments.")
		os.Exit(1)
	}

	switch args[0] {
	case "flash":
		if len(args) == 1 {
			flash_all(green)
		} else if len(args) == 2 {
			flash_all(args[1])
		}
	case "all":
		if len(args) == 2 {
			all_on(args[1])
		} else {
			all_on(green)
		}
	case "on":
		if len(args) == 2 {
			led_on(args[1], args[2])
		} else {
			led_on(args[1], green)
		}
	case "off":
		if len(args) == 1 {
			all_off()
		} else {
			led_off(args[1])
		}
	case "listen":
		listen()

	case "draw":
		draw()

	default:
		fmt.Println("INVALID USAGE")
	}

}

// function to flash all leds
func flash_all(color string) {
	for {
		all_on(color)
		time.Sleep(time.Second)
		all_off()
		time.Sleep(time.Second)
	}
}

// function to constantly monitor launchapd input
func listen() error {
	for {
		output, err := get_btn()
		if err != nil {
			return fmt.Errorf("Error listening: %v", err)
		}
		fmt.Println(string(output))
	}
}

// function to get one line of launchpad input
func get_btn() (string, error) {
	cmd := exec.Command(lpCmd, getArgs...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("Error creating stdout: %v", err)
	}
	cmd.Start()

	var output = make([]byte, 9)
	n := 0
	for n < 9 {
		n, err = stdout.Read(output)
		if err != nil {
			return "", fmt.Errorf("Error reading stdout: %v", err)
		}
	}
	cmd.Process.Kill()
	return string(output), nil
}

// function to turn led at x,y on to specified color
func led_on(pos string, color string) {
	args := append(pushArgs, fmt.Sprintf("90 %s %s", pos, color))
	cmd := exec.Command(lpCmd, args...)
	cmd.Run()

}

// function to turn off led at x,y
func led_off(pos string) {
	args := append(pushArgs, fmt.Sprintf("90 %s 00", pos))
	cmd := exec.Command(lpCmd, args...)
	cmd.Run()
}

// function to turn of led at x,y on the top row
func top_on(pos string, color string) {
	args := append(pushArgs, fmt.Sprintf("B0 %s %s", pos, color))
	cmd := exec.Command(lpCmd, args...)
	cmd.Run()
}

// function to turn of led at x,y on the top row
func top_off(pos string) {
	args := append(pushArgs, fmt.Sprintf("B0 %s 00", pos))
	cmd := exec.Command(lpCmd, args...)
	cmd.Run()
}

// function to turn all leds on to specified color
func all_on(color string) {
	for i := range 8 {
		top_on(fmt.Sprintf("6%x", 8+i), color)
		for j := range 9 {
			pos := fmt.Sprintf("%d%d", i, j)
			led_on(pos, color)
		}
	}
}

// function to turn all leds on to specified color
func all_off() {
	for i := range 8 {
		top_off(fmt.Sprintf("6%x", 8+i))
		for j := range 9 {
			pos := fmt.Sprintf("%d%d", i, j)
			led_off(pos)
		}
	}
}

// function to turn on any pushed leds
func draw() error {
	for {
		output, err := get_btn()
		if err != nil {
			return fmt.Errorf("Error listening: %v", err)
		}
		pos := strings.Split(output, " ")[1]
		led_on(pos, "30")
	}
}
