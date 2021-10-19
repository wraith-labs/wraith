package libwraith

import "time"

type TxQueueElement struct {
	Addr                  string
	Encoding              string
	TransmissionFailCount int
	TransmissionFailTime  time.Time
	Data                  map[string]interface{}
}
