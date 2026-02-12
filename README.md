# Programmable Novation Launchpad S 
## About
* An application written in Go for using the Novation Launchpad S as a programmable macro pad.
* Created because I wanted something similar to an Elgato Streamdeck but already had a Novation Launchpad.

## Requirements
* Novation Launchpad S
* Go version 1.25.5 or newer
* amidi version 1.2.15.2 or newer

## Setup
* clone this repository with `git clone https://github.com/TaitA2/Launchpad.git`
* build the program with `go build .`
* execute the program called `launchpad`

## Usage
* The program includes 8 layers, controlled by the top row of buttons
  * pressing the button matching the current layer will refresh the grid LEDs
* The global color theme can be set by selecting any of the 8 colorful right column buttons
* The main grid of buttons are the main point of interaction

### Layers
0. Freeze          - Pressing a grid button lights it the selected color until released.
1. Paint           - Pressing a grid button lights it the selected color until pressed again with a new color.
2. Breathe         - Flashes the grid as the selected color originating from the center.
3. All on          - Enables all grid LEDs as the selected color.
4. Macro           - Grid buttons with an existing macro binding will be lit. Pressing the button will perform the assigned macro.
5. Macro recording - Pressing a grid button prompts the user for input. The command entered is saved to the button pressed. (Entering no command will clear the command for that button).
6. Color debug     - Displays all possible LED colors. Will be used for further color customisation in future.
7. Unimplemented   - Will be game of life in future.
