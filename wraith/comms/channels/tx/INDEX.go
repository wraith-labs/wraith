package tx

// The Tx struct - a template for transmitters
type Tx struct {
	Start   func()
	Stop    func()
	Trigger func(data TxQueueElement) bool
	Data    map[string]interface{}
}

// Stuff for storing data about to be sent
type TxQueue chan TxQueueElement
type TxQueueElement struct {
	Addr string
	Data map[string]interface{}
}

var UnifiedTxQueue TxQueue

func init() {
	UnifiedTxQueue = make(TxQueue)
}
