package main

import (
	"fmt"
	"github.com/stianeikeland/go-rpio"
	"os"
	"time"
)

var (
	pin = rpio.Pin(0)
)

func flashLight() {
	//Open and map memory to access gpio, check for erros
	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Unmap gpio memory when done
	defer rpio.Close()

	// Set pin to output mode
	pin.Output()

	// Toggle - infinite loop
	for {
		pin.Toggle()
		time.Sleep(time.Second)
	}
}
