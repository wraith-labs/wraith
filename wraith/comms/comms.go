package comms

import (
	"net/url"

	"github.com/0x1a8510f2/wraith/hooks"
)

// Structures (templates) for transmitters and receivers
type Tx struct {
	Start   func()
	Stop    func()
	Trigger func(data TxQueueElement) bool
	Data    map[string]interface{}
}
type Rx struct {
	Start func()
	Stop  func()
	Data  map[string]interface{}
}

// The queue types, which store inbound and outbound data
type TxQueue chan TxQueueElement
type TxQueueElement struct {
	Addr string
	Data map[string]interface{}
}
type RxQueue chan RxQueueElement
type RxQueueElement struct {
	Data map[string]interface{}
}

// Maps mapping URL schemes to individual transmitters and receivers
var transmitters map[string]*Tx
var receivers map[string]*Rx

// Unified queues to hold incoming and outgoing data
var UnifiedTxQueue TxQueue
var UnifiedRxQueue RxQueue

// Channel used to make the comms manager exit cleanly
var managerExitTrigger chan struct{}

// Register a transmitter to make it useable by the Wraith
func RegTx(scheme string, tx *Tx) {
	// Initialise the data map
	tx.Data = make(map[string]interface{})

	transmitters[scheme] = tx
}

// Register a receiver to make it useable by the Wraith (and inject the unifiedRxQueue)
func RegRx(scheme string, rx *Rx) {
	// Initialise the data map
	rx.Data = make(map[string]interface{})
	// Inject the unified queue
	rx.Data["queue"] = UnifiedRxQueue

	receivers[scheme] = rx
}

// Infinite loop managing data transmission
// This should run in a thread and only a single instance should run at a time
func Manage() {
	// Always stop transmitters and receivers before exiting
	defer func() {
		for _, transmitter := range transmitters {
			transmitter.Stop()
		}
		for _, receiver := range receivers {
			receiver.Stop()
		}
		close(managerExitTrigger)
	}()

	// Start transmitters and receivers
	for _, transmitter := range transmitters {
		transmitter.Start()
	}
	for _, receiver := range receivers {
		receiver.Start()
	}

	for {
		select {
		case <-managerExitTrigger:
			return
		case tx := <-UnifiedTxQueue:
			txaddr, err := url.Parse(tx.Addr)
			// If there was an error parsing the URL, the whole txdata should be dropped as there's nothing more we can do
			if err == nil {
				// ...same in case of a non-existent transmitter
				if transmitter, exists := transmitters[txaddr.Scheme]; exists {
					transmitter.Trigger(tx)
				}
			}
		}
	}
}

func init() {
	// Initialise variables
	transmitters = make(map[string]*Tx)
	receivers = make(map[string]*Rx)
	UnifiedTxQueue = make(TxQueue)
	UnifiedRxQueue = make(RxQueue)
	managerExitTrigger = make(chan struct{})

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
