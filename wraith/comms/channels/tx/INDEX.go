package tx

// The Tx struct - a template for transmitters
type Tx struct {
	Start   func()
	Stop    func()
	Trigger func(data TxQueueElement) bool
	Data    map[string]interface{}
}

// List of transmitters (Tx)
type txList struct {
	data map[string]*Tx
}

func (txl *txList) InitIfNot() {
	if txl.data == nil {
		txl.data = make(map[string]*Tx)
	}
}
func (txl *txList) Add(scheme string, handler *Tx) {
	txl.InitIfNot()

	// Initialise the handler
	handler.Data = make(map[string]interface{})

	txl.data[scheme] = handler
}
func (txl *txList) Get(scheme string) (*Tx, bool) {
	txl.InitIfNot()

	tx, ok := txl.data[scheme]
	return tx, ok
}
func (txl *txList) GetList() map[string]*Tx {
	txl.InitIfNot()

	return txl.data
}

// Stuff for storing data about to be sent
type TxQueue chan TxQueueElement
type TxQueueElement struct {
	Addr string
	Data map[string]interface{}
}

var UnifiedTxQueue TxQueue
var TxList txList

func init() {
	UnifiedTxQueue = make(TxQueue)
}
