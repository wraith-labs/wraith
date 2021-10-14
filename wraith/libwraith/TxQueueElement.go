package libwraith

type TxQueueElement struct {
	Addr     string
	Encoding string
	Data     map[string]interface{}
}
