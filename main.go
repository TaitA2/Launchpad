package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var lpCmd string = "amidi"
var getArgs = []string{"-d", "-p", "hw:0,0,0"}
var pushArgs = []string{"-p", "hw:0,0,0", "-S"}

// top row vs grid row codes
const topRow = "B0"
const gridRow = "90"

// LED color codes
const off = 0
const green = 30
const red = 3
const amber = 13
const lime = 39

// button struct
type button struct {
	row     string // topRow or gridRow
	x       int    // collumn index
	y       int    // row index
	color   int    // current button color
	pressed bool   // currently held down
}

// launchpad struct
type launchpad struct {
	topButtons  []*button   // map of x index to topRow buttons
	gridButtons [][]*button // 2D array of buttons - first index for row, second index for collumn
	layer       int         // current active 'layer' (0-7) tied to top row
}

func main() {
	all_off()
	var lp *launchpad
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("Too few arguments.")
		os.Exit(1)
	}

	switch args[0] {
	case "test":
		push_test()
	case "pallette":
		pallette()
	case "flash":
		if len(args) == 1 {
			lp.flash_all(green)
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
			lp.grid_on(args[1], args[2])
		} else {
			lp.grid_on(args[1], green)
		}
	case "off":
		if len(args) == 1 {
			lp.all_off()
		} else {
			led_off(args[1])
		}
	case "listen":
		listen()

	case "paint":
		paint()

	default:
		fmt.Println("INVALID USAGE")
	}

}

// function to flash all leds
func (lp *launchpad) flash_all(color int) {
	for {
		lp.all_on(color)
		time.Sleep(time.Second)
		lp.all_off()
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
func (b *button) led_on(color int) {
	args := append(pushArgs, fmt.Sprintf("%s %d%d %s", b.row, b.x, b.y, color))
	cmd := exec.Command(lpCmd, args...)
	cmd.Run()

}

// function to turn off led at x,y
func (b *button) led_off() {
	args := append(pushArgs, fmt.Sprintf("%s %d%d 00", b.row, b.x, b.y))
	cmd := exec.Command(lpCmd, args...)
	cmd.Run()
}

// function to turn all leds on to specified color
func (lp *launchpad) all_on(color int) {

	// turn on all top buttons

	for _, btn := range lp.topButtons {
		btn.led_on(color)
	}

	// turn on all grid buttons
	for _, row := range lp.gridButtons {
		for _, btn := range row {
			btn.led_on(color)
		}
	}
}

// function to turn all leds on to specified color
func (lp *launchpad) all_off() {
	// turn off all top buttons
	for _, btn := range lp.topButtons {
		btn.led_off()
	}

	// turn off all grid buttons
	for _, row := range lp.gridButtons {
		for _, btn := range row {
			btn.led_off()
		}
	}
}

// function to turn on any pushed leds
func paint() error {
	paint_setup()
	for {
		output, err := get_btn()
		if err != nil {
			return fmt.Errorf("Error listening: %v", err)
		}
		pos := strings.Split(output, " ")[1]
		grid_on(pos, green)
	}
}

func paint_setup() {
	all_off()

	// right collumn color pallete
	grid_on("90 08", green)
}

// function to enable LED of any button while its pushed
func push_test() error {
	for {
		output, err := get_btn()
		if err != nil {
			return fmt.Errorf("Error listening: %v", err)
		}
		outsplit := strings.Split(output, " ")
		color := outsplit[2]
		pos := strings.Join(outsplit[:2], " ")
		grid_on(pos, color)
	}
}

func pallette() {
	fmt.Println("Filling colors...")
	k := 0
	for {
		start := k
		for i := range 8 {
			for j := range 4 {
				pos := fmt.Sprintf("%d%d", i, j)
				color := strconv.Itoa(k)
				grid_on(pos, color)
				k++
			}
		}
		fmt.Printf("Showing %d - %d\n", start, k)
		for i := range 8 {
			for j := range 4 {
				pos := fmt.Sprintf("%d%d", i, j+4)
				color := strconv.Itoa(k)
				grid_on(pos, color)
				k++
			}
		}
		time.Sleep(time.Second * 5)
	}
}
