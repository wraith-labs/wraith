package comms

import (
	"net/url"
	"sync"
)

// Structures for transmitters and receivers
type Tx struct {
	Start   func()
	Stop    func()
	Main    func()
	Trigger func()
	Queue   TxQueue
}
type Rx struct {
	Start   func()
	Stop    func()
	Main    func()
	Trigger func()
	Queue   RxQueue
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

// A unified rx channel to collect data from all receivers in one place
var unifiedRxQueue RxQueue

// Channel used to make the comms manager exit cleanly
var managerExitTrigger chan struct{}

// Register a transmitter to make it useable by the Wraith
func RegTx(scheme string, tx *Tx) {
	transmitters[scheme] = tx
}

// Register a receiver to make it useable by the Wraith (and inject the unifiedRxQueue)
func RegRx(scheme string, rx *Rx) {
	rx.Queue = unifiedRxQueue
	receivers[scheme] = rx
}

// Infinite loop managing transmission and receiving of data
func Manage(txQueue TxQueue, rxQueue RxQueue, wg *sync.WaitGroup) {
	// Make sure the waitgroup is always decremented when this function exits
	defer func() {
		wg.Done()
	}()

	for {
		select {
		case <-managerExitTrigger:
			return
		case tx := <-txQueue:
			txaddr, err := url.Parse(tx.Addr)
			// If there was an error parsing the URL, the whole txdata should be dropped as there's nothing more we can do
			if err == nil {
				// ...same in case of a non-existent transmitter
				if transmitter, exists := transmitters[txaddr.Scheme]; exists {
					transmitter.Queue <- tx
				}
			}
		case rxQueue <- (<-unifiedRxQueue):
		}
	}
}

func init() {
	// Initialise channel lists
	transmitters = make(map[string]*Tx)
	receivers = make(map[string]*Rx)
}
