package libwraith

// Reserved locations in the shared memory with special purposes.
// All other locations should be namespaced.
const (
	// This cell holds the latest error which occurred, be it in a module
	// or Wraith itself. Can be used to send error logs to C2.
	SHM_ERRS = "err"
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
