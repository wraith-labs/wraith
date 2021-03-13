package main

import (
	"time"

	"github.com/TR-SLimey/wraith/hooks"
	"github.com/TR-SLimey/wraith/radio"
)

// Useful globals
var startTime time.Time

// Main wraith struct
type wraith struct {
	ID string
}

func main() {
	// Run OnStart hooks
	hooks.RunOnStart()

	r := radio.NewRadio()
	go r.RunTransmit()
	<-time.After(20 * time.Second)
}
