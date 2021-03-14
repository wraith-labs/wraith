package main

import (
	"fmt"
	"sync"
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

	// Run OnExit hooks always on exit
	defer hooks.RunOnExit()

	// Keep track of active goroutines
	var wg sync.WaitGroup

	txQueue := make(comms.TxQueue)
	rxQueue := make(comms.RxQueue)
	wg.Add(1)
	go comms.Manage(txQueue, rxQueue, &wg)

	// Mainloop: Transmit, receive and process stuff
	for {
		select {
		case data := <-rxQueue:
			fmt.Printf("%v", data)
		}
	}

}
