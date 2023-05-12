package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stianeikeland/go-rpio"
	"net/http"
	"os"
	"time"
)

// add more pins if needed here
// take care of the GPIO <-> PIN mapping
var (
	pin = rpio.Pin(0) //Pin 27
)

// Just for testing
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

// Inits the pin
func initPin() {
	//Open and map memory to access gpio, check for errors
	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Unmap gpio memory when done
	defer rpio.Close()

	// Set pin to output mode
	pin.Output()
	fmt.Println("Pins successfully initialized")
}

func updatePinLocal(b bool) {
	if b {
		pin.High()
		fmt.Println("Pin turned on")
	} else {
		pin.Low()
		fmt.Println("Pin turned off")
	}
}

func updatePinLocalDummy(b bool) {
	if b {
		fmt.Println("Pin turned on")
	} else {
		fmt.Println("Pin turned off")
	}
}

// Turns the gpio on or off according to the current @param currentRequest (useage with api)
func updatePin(c *gin.Context) {
	var temp bool

	if err := c.Bind(&temp); err != nil {
		fmt.Println("Failed to load the data into variable %b", temp)
		return
	}

	if temp {
		pin.High()
		fmt.Println("Pin turned on")
	} else {
		pin.Low()
		fmt.Println("Pin turned off")
	}
	c.IndentedJSON(http.StatusAccepted, temp)
}
