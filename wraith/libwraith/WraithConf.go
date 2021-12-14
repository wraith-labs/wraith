package libwraith

import "time"

type WraithConf struct {
	FamilyId              string
	FingerprintGenerator  func() string
	DefaultReturnAddr     string
	DefaultReturnEncoding string
	RetransmissionDelay   time.Duration
	RetransmissionCap     int
	Custom                map[string]interface{}
}
