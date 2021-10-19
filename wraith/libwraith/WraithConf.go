package libwraith

import "time"

type WraithConf struct {
	Fingerprint           string
	DefaultReturnAddr     string
	DefaultReturnEncoding string
	RetransmissionDelay   time.Duration
	RetransmissionCap     int
	Custom                map[string]interface{}
}
