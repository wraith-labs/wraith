package main

import (
	"time"

	"github.com/0x1a8510f2/wraith/comms"
	"github.com/0x1a8510f2/wraith/hooks"
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

	rcvchan := make(chan map[string]interface{})
	comms.Manage(&rcvchan)

	// Run OnExit hooks
	hooks.RunOnExit()
}
