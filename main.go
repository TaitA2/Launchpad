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

// map of string color names to color codes
var colors = map[string]int{"green": green, "red": red, "amber": amber, "lime": lime}

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
	topButtons   []*button   // array x index to topRow buttons
	rightButtons []*button   // array y index of right collumn buttons
	gridButtons  [][]*button // 2D array of buttons - first index for row, second index for collumn
	layer        int         // current active 'layer' (0-7) tied to top row
}

func main() {
	// get the launchpad struct
	lp := get_launchpad()
	// turn off all LEDs
	lp.all_off()

	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("Too few arguments.")
		os.Exit(1)
	}

	switch args[0] {
	case "test":
		lp.push_test()
	case "pallette":
		lp.pallette()
	case "flash":
		if len(args) == 1 {
			lp.flash_all(green)
		} else if len(args) == 2 {
			lp.flash_all(colors[args[1]])
		}
	case "all":
		if len(args) == 2 {
			lp.all_on(colors[args[1]])
		} else {
			lp.all_on(green)
		}
	case "on":
		if len(args) == 2 {
			lp.grid_on(colors[args[1]])
		} else {
			lp.grid_on(green)
		}
	case "off":
		lp.all_off()
	case "listen":
		lp.listen()

	case "paint":
		lp.paint()

	default:
		fmt.Println("INVALID USAGE")
	}

}

// function to return launchpad struct
func get_launchpad() *launchpad {
	var lp launchpad
	lp.topButtons = make([]*button, 8)
	lp.rightButtons = make([]*button, 8)
	lp.gridButtons = make([][]*button, 8)
	for i := range 8 {
		lp.gridButtons[i] = make([]*button, 8)
		topBtn := button{row: topRow, x: i, y: i, color: off, pressed: false}
		lp.topButtons[i] = &topBtn
		rightBtn := button{row: gridRow, x: 8, y: i, color: off, pressed: false}
		lp.rightButtons[i] = &rightBtn
		for j := range 8 {
			gridBtn := button{row: gridRow, x: j, y: i, color: off, pressed: false}
			lp.gridButtons[i][j] = &gridBtn
		}
	}
	return &lp
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
func (lp *launchpad) listen() error {
	for {
		b, err := lp.get_btn()
		if err != nil {
			return fmt.Errorf("Error listening: %v", err)
		}
		fmt.Printf("%v\n", b)
	}
}

// function to get one line of launchpad input
func (lp *launchpad) get_btn() (*button, error) {
	cmd := exec.Command(lpCmd, getArgs...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("Error creating stdout: %v", err)
	}
	cmd.Start()
	defer cmd.Process.Kill()

	var output = make([]byte, 9)
	n := 0
	for n < 9 {
		n, err = stdout.Read(output)
		if err != nil {
			return nil, fmt.Errorf("Error reading stdout: %v", err)
		}
	}
	parts := strings.Split(string(output), " ")

	var b *button
	row := parts[0]
	x, _ := strconv.Atoi(parts[1][0:1])
	y, _ := strconv.ParseInt(parts[1][1:], 16, 64)
	pressed := parts[2] != "00"

	if strings.Contains(row, topRow) {
		b = lp.topButtons[y-8]
	} else if y == 8 {
		b = lp.rightButtons[x]

	} else {
		b = lp.gridButtons[x][y]
	}
	b.pressed = pressed

	return b, nil
}

// function to turn led at x,y on to specified color
func (b *button) led_on(color int) {
	b.color = color
	args := append(pushArgs, fmt.Sprintf("%s %d%d %d", b.row, b.x, b.y, color))
	cmd := exec.Command(lpCmd, args...)
	cmd.Run()

}

// function to turn off led at x,y
func (b *button) led_off() {
	args := append(pushArgs, fmt.Sprintf("%s %d%d 00", b.row, b.x, b.y))
	cmd := exec.Command(lpCmd, args...)
	cmd.Run()
}

func (lp *launchpad) grid_on(color int) {

	// turn on all grid buttons
	for _, row := range lp.gridButtons {
		for _, btn := range row {
			btn.led_on(color)
		}
	}

}

// function to turn all leds on to specified color
func (lp *launchpad) all_on(color int) {

	// turn on all top buttons

	for _, btn := range lp.topButtons {
		btn.led_on(color)
	}

	lp.grid_on(color)
}

// function to turn all leds on to specified color
func (lp *launchpad) all_off() {
	// turn off all top buttons
	for i := range lp.topButtons {
		lp.topButtons[i].led_off()
	}

	// turn off all grid buttons
	for i, row := range lp.gridButtons {
		for j := range row {
			lp.gridButtons[i][j].led_off()
		}
	}
}

// function to turn on any pushed leds
func (lp *launchpad) paint() error {
	lp.paint_setup()
	for {
		b, err := lp.get_btn()
		if err != nil {
			return fmt.Errorf("Error listening: %v", err)
		}
		b.led_on(green)
	}
}

func (lp *launchpad) paint_setup() {
	lp.all_off()

}

// function to enable LED of any button while its pushed
func (lp *launchpad) push_test() error {
	for {
		b, err := lp.get_btn()
		if err != nil {
			return fmt.Errorf("Error listening: %v", err)
		}
		if b.pressed {
			b.led_on(green)
		} else {
			b.led_off()
		}
	}
}

func (lp *launchpad) pallette() {
	fmt.Println("Filling colors...")
	k := 0
	for {
		start := k
		for i := range 8 {
			for j := range 4 {
				lp.gridButtons[i][j].led_on(k)
				k++
			}
		}
		fmt.Printf("Showing %d - %d\n", start, k)
		for i := range 8 {
			for j := range 4 {
				lp.gridButtons[i][j+4].led_on(k)
				k++
			}
		}
		time.Sleep(time.Second * 5)
	}
}
