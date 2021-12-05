package libwraith

// Reserved locations in the shared memory with special purposes
// All other locations should be namespaced
const (
	// This cell forces Wraith to exit when it is written to. The value
	// is irrelevant.
	SHM_EXIT_TRIGGER = "exitTrigger"

	// This cell forces Wraith to stop and restart all modules whenever
	// it is written to. The value is irrelevant.
	SHM_RELOAD_TRIGGER = "reloadTrigger"

	// This cell stores all data which is to be transmitted to C2 in the
	// form of a channel. This data should be managed and directed to
	// individual comms modules by the MOD_COMMS_MANAGER module.
	SHM_TX_QUEUE = "txQueue"

	// This cell stores all data which has been received from C2 and is
	// awaiting processing, in the form of a channel. This data should
	// be managed and directed to individual comms modules by the
	// MOD_COMMS_MANAGER module.
	SHM_RX_QUEUE = "rxQueue"
)

// Reserved module names for modules with special purposes
// All other modules should be namespaced
const (
	// This module is responsible for managing the SHM_TX_QUEUE and
	// SHM_RX_QUEUE memory cells and distributing the data within them
	// to individual modules responsible for its handling.
	MOD_COMMS_MANAGER = "commsManager"
)
