// +build debug

/*
This is a debug receiver for testing purposes. It only gets included in debug builds.
When included, it "receives" a print command every 2 seconds.
*/

package rx

import "github.com/0x1a8510f2/wraith/comms"

var Debug comms.Rx

func init() {
	Debug.Start = func() { println(1) }
	Debug.Stop = func() {}
	comms.RegRx("debug", &Debug)
}
