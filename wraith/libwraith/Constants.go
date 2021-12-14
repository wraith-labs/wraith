package libwraith

// Wraith status codes for use in SHM_WRAITH_STATUS
const (
	// Wraith is not running
	WSTATUS_INACTIVE = iota

	// Wraith is running correctly
	WSTATUS_ACTIVE

	// Wraith is shutting down
	WSTATUS_DEACTIVATING

	// Wraith has exited with a fatal error
	WSTATUS_ERROR
)

// Reserved locations in the shared memory with special purposes
// All other locations should be namespaced
const (
	// This cell forces Wraith to stop and restart all modules whenever
	// it is written to. The value is irrelevant. It can be written to
	// by any component.
	SHM_RELOAD_TRIGGER = "reloadTrigger"

	// This cell contains the current status of Wraith as defined by the
	// STATUS_ constants. It MUST NOT be written to by any component other
	// than Wraith itself.
	SHM_WRAITH_STATUS = "status"

	// This cell stores data which is to be transmitted to C2. This data
	// should be managed and directed to individual comms modules by the
	// MOD_COMMS_MANAGER module, hence only that module should really read
	// from this cell, but any module can write to it.
	SHM_TX_QUEUE = "txQueue"

	// This cell stores all data which has been received from C2 and is
	// awaiting processing. This data should be managed and directed to
	// individual comms modules by the MOD_COMMS_MANAGER module. Hence,
	// only that module should really read from this cell, but any module
	// can write to it.
	SHM_RX_QUEUE = "rxQueue"

	// This cell holds the status of the CommsManager. Modules should not
	// write to the TX/RX queue cells if this value is falsey or nil as
	// their messages can get lost.
	SHM_COMMS_READY = "commsReady"
)

// Reserved module names for modules with special purposes
// All other modules should be namespaced
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
	MOD_COMMS_MANAGER = "commsManager"
)

// Configuration for SharedMemory
const (
	// The size of watcher channels. Making this bigger makes update
	// delivery more reliable and ordered but increases memory usage
	// if a watcher isn't reading its updates.
	SHMCONF_WATCHER_CHAN_SIZE = 255

	// Timeout in seconds after which notifications for watchers are
	// dropped if writing to the channel blocks.
	SHMCONF_WATCHER_NOTIF_TIMEOUT = 1
)
