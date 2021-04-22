package rx

// The Rx struct - a template for receivers
type Rx struct {
	Start func()
	Stop  func()
	Data  map[string]interface{}
}

// List of receivers (Rx)
type rxList struct {
	data map[string]*Rx
}

func (rxl *rxList) InitIfNot() {
	if rxl.data == nil {
		rxl.data = make(map[string]*Rx)
	}
}
func (rxl *rxList) Add(scheme string, handler *Rx) {
	rxl.InitIfNot()

	// Initialise the handler
	handler.Data = make(map[string]interface{})
	handler.Data["queue"] = UnifiedRxQueue

	rxl.data[scheme] = handler
}
func (rxl *rxList) Get(scheme string) (*Rx, bool) {
	rxl.InitIfNot()

	rx, ok := rxl.data[scheme]
	return rx, ok
}
func (rxl *rxList) GetList() map[string]*Rx {
	rxl.InitIfNot()

	return rxl.data
}

// Stuff for storing just-received data
type RxQueue chan RxQueueElement
type RxQueueElement struct {
	Data []byte
}

var UnifiedRxQueue RxQueue
var RxList rxList

func init() {
	UnifiedRxQueue = make(RxQueue)
}
