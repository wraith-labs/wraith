package comms

import (
	"net/url"

	"github.com/0x1a8510f2/wraith/hooks"

	"github.com/0x1a8510f2/wraith/comms/channels/rx"
	"github.com/0x1a8510f2/wraith/comms/channels/tx"
)

// Channel used to make the comms manager exit cleanly
var managerExitTrigger chan struct{}

// Passthrough from tx and rx modules
var UnifiedTxQueue *tx.TxQueue
var UnifiedRxQueue *rx.RxQueue

// Infinite loop managing data transmission
// This should run in a thread and only a single instance should run at a time
func Manage() {
	// Always stop transmitters and receivers before exiting
	defer func() {
		for _, transmitter := range tx.TxList.GetList() {
			transmitter.Stop()
		}
		for _, receiver := range rx.RxList.GetList() {
			receiver.Stop()
		}
		close(managerExitTrigger)
	}()

	// Start transmitters and receivers
	for _, transmitter := range tx.TxList.GetList() {
		transmitter.Start()
	}
	for _, receiver := range rx.RxList.GetList() {
		receiver.Start()
	}

	for {
		select {
		case <-managerExitTrigger:
			return
		case transmission := <-tx.UnifiedTxQueue:
			txaddr, err := url.Parse(transmission.Addr)
			// If there was an error parsing the URL, the whole txdata should be dropped as there's nothing more we can do
			if err == nil {
				// ...same in case of a non-existent transmitter
				if transmitter, exists := tx.TxList.Get(txaddr.Scheme); exists {
					transmitter.Trigger(transmission)
				}
			}
		}
	}
}

func init() {
	// Initialise variables
	managerExitTrigger = make(chan struct{})
	UnifiedTxQueue = &tx.UnifiedTxQueue
	UnifiedRxQueue = &rx.UnifiedRxQueue

	// Hook the comms manager into the on start and on exit events
	hooks.OnStart.Add(func() {
		go Manage()
	})
	hooks.OnExit.Add(func() {
		// Trigger exit
		managerExitTrigger <- struct{}{}
		// Wait until channel is closed (exit confirmed)
		<-managerExitTrigger
	})
}
