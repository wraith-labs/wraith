package libwraith

import "time"

// A struct providing configuration options for Wraith to allow
// for altering behaviour without altering the code.
type Config struct {
	// A string representing the family ID or strain ID of Wraith.
	// This can be useful to check what different versions of
	// Wraith are out there, or to target only one specific
	// strain with commands/payloads. This should be changed
	// whenever a significant change is made to Wraith before building.
	StrainId string

	// A function used to generate the fingerprint for this instance
	// of Wraith. That is, a unique string identifying specifically this
	// binary, on this host, in this process. It can be a UUID, for
	// instance, meaning that it serves only the purpose of identifiaction
	// and changes on every Wraith restart, or a string based on some
	// information such as MAC Address+Wraith PID.
	FingerprintGenerator func() string

	// The max time to wait for a heartbeat from Wraith's mainloop before
	// assuming that this instance is dead. Around 1 second is recommended.
	// Note that setting this too high can cause significant slowdowns when
	// Wraith does die.
	HeartbeatTimeout time.Duration

	// How many times modules should be allowed to crash within a time
	// specified in ModuleCrashLoopDetectTime before they are no longer
	// restarted. It is recommended to keep this relatively low to prevent
	// buggy modules from using up resources. The lower the value the more
	// strict the crashloop detection.
	ModuleCrashloopDetectCount int

	// After this time, module crashes are forgotten when evaluating whether
	// a module is crashlooping. It is recommended to keep this value relatively
	// high to ensure that crashlooped or buggy modules are always caught. The
	// higher the value the more strict the crashloop detection.
	ModuleCrashloopDetectTime time.Duration
}
