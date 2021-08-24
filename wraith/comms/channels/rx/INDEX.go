package rx

// The Rx struct - a template for receivers
type Rx struct {
	Start func()
	Stop  func()
	Data  map[string]interface{}
}

// Stuff for storing just-received data
type RxQueue chan RxQueueElement
type RxQueueElement struct {
	Data []byte
}

var UnifiedRxQueue RxQueue

func init() {
	UnifiedRxQueue = make(RxQueue)
}
