#!/bin/bash

# Set your device here
DEVICE="hw:0,0,0"  # Replace with your device from amidi -l

# Function: Light LED at note with velocity
led_on() {
    local note=$1
    local vel=$2
    amidi -p "$DEVICE" -S "90 $note $vel"
}

# Function: Turn off LED
led_off() {
    local note=$1
    amidi -p "$DEVICE" -S "80 $note 00"
}

# Function: Light all LEDs
all_on() {
    for note in {0..80}; do
        led_on $note 39
        # sleep 0.02
    done
}

# Function: Turn all off
all_off() {
    for note in {0..80}; do
        led_off $note
    done
}

# Function: Light grid row by row
rainbow_grid() {
    for row in {0..8}; do
        for col in {0..7}; do
            note=$((row * 8 + col))
            vel=$(( (row * 16) + (col * 2) ))
            if [ $vel -gt 127 ]; then vel=127; fi
            led_on $note $vel
        done
        sleep 0.1
    done
}

led_test() {
    for vel in {0..1000}; do
        led_on 00 $vel
        echo $vel
        sleep 1
    done
}

smile() {
    color=30
    all_off
    led_on 22 $color
    led_on 25 $color
    led_on 41 $color
    led_on 52 $color
    led_on 53 $color
    led_on 54 $color
    led_on 55 $color
    led_on 46 $color

}

flash_smile() {
    for i in {0..5}; do
        smile
        sleep 0.1
        all_off
        sleep 0.1
    done
}

draw() {
    all_off
    while true; do
        led=$(amidi -d -p hw:0,0,0 | grep 90 -m 1 | grep " .. " -o)
        led_on $led 30

    done
}
