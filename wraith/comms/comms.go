package comms

import (
	"net/url"

	"git.0x1a8510f2.space/0x1a8510f2/wraith/hooks"
	mm "git.0x1a8510f2.space/0x1a8510f2/wraith/modmgr"
	"git.0x1a8510f2.space/0x1a8510f2/wraith/types"
)

// Channel used to make the comms manager exit cleanly
var managerExitTrigger chan struct{}

// Inbound and outbound queues
var UnifiedTxQueue types.TxQueue
var UnifiedRxQueue types.RxQueue

// Infinite loop managing data transmission
// This should run in a thread and only a single instance should run at a time
func Manage() {
	// Always stop transmitters and receivers before exiting
	defer func() {
		if transmitters, ok := mm.Modules.GetEnabled(mm.ModCommsChanTx).(map[string]mm.CommsChanTxModule); ok {
			for _, transmitter := range transmitters {
				transmitter.StopTx()
			}
		}
		if receivers, ok := mm.Modules.GetEnabled(mm.ModCommsChanRx).(map[string]mm.CommsChanRxModule); ok {
			for _, receiver := range receivers {
				receiver.StopRx()
			}
		}
		close(managerExitTrigger)
	}()

	// Start transmitters and receivers
	if transmitters, ok := mm.Modules.GetEnabled(mm.ModCommsChanTx).(map[string]mm.CommsChanTxModule); ok {
		for _, transmitter := range transmitters {
			transmitter.StartTx()
		}
	}
	if receivers, ok := mm.Modules.GetEnabled(mm.ModCommsChanRx).(map[string]mm.CommsChanRxModule); ok {
		for _, receiver := range receivers {
			receiver.StartRx()
		}
	}

	for {
		select {
		case <-managerExitTrigger:
			return
		case transmission := <-UnifiedTxQueue:
			txaddr, err := url.Parse(transmission.Addr)
			// If there was an error parsing the URL, the whole txdata should be dropped as there's nothing more we can do
			if err == nil {
				// ...same in case of a non-existent transmitter
				if transmitters, ok := mm.Modules.GetEnabled(mm.ModCommsChanTx).(map[string]mm.CommsChanTxModule); ok {
					if transmitter, exists := transmitters[txaddr.Scheme]; exists {
						transmitter.TriggerTx(transmission)
					}
				}
			}
		}
	}
}

func init() {
	// Initialise variables
	managerExitTrigger = make(chan struct{})
	UnifiedTxQueue = make(types.TxQueue)
	UnifiedRxQueue = make(types.RxQueue)

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
