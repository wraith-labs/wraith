package comms

import (
	"net/url"

	"github.com/0x1a8510f2/wraith/hooks"
)

// Structures (templates) for transmitters and receivers
type Tx struct {
	Start func()
	Stop  func()
	Main  func()
	Data  map[interface{}]interface{}
}
type Rx struct {
	Start func()
	Stop  func()
	Main  func()
	Data  map[interface{}]interface{}
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
	transmitters[scheme] = tx
}

// Register a receiver to make it useable by the Wraith (and inject the unifiedRxQueue)
func RegRx(scheme string, rx *Rx) {
	rx.Queue = UnifiedRxQueue
	receivers[scheme] = rx
}

// Infinite loop managing data transmission
func Manage() {
	// Initialise unified queues
	UnifiedTxQueue = make(TxQueue)
	UnifiedRxQueue = make(RxQueue)

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
					transmitter.Queue <- tx
				}
			}
		}
	}
}

func init() {
	// Initialise channel lists
	transmitters = make(map[string]*Tx)
	receivers = make(map[string]*Rx)
	// Hook the comms manager into the on start and on exit events
	hooks.OnStart.Add(func() {
		go Manage()
	})
}
