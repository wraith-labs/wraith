package main

import (
	"time"

	"github.com/TR-SLimey/wraith/comms"
	"github.com/TR-SLimey/wraith/hooks"
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

	r := comms.NewRadio()
	go r.RunTransmit()
	<-time.After(20 * time.Second)
}
