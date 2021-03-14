// +build debug

/*
This is a debug receiver for testing purposes. It only gets included in debug builds.
When included, it "receives" a print command every 2 seconds.
*/

package receivers

import "github.com/0x1a8510f2/wraith/comms"

var Debug comms.Rx

func init() {
	comms.RegRx("debug", &Debug)
}
