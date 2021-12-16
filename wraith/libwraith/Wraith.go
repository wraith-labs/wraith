package libwraith

import (
	"fmt"
	"sync"
	"time"
)

type Wraith struct {
	// Keep track of the time Wraith was initialised/started so it
	// can be retrieved by modules if needed.
	initTime time.Time

	// A fingerprint to uniquely identify this Wraith. It is
	// generated on init. This helps to target individual Wraiths
	// with commands, for instance.
	fingerprint string

	// An instance of the SharedMemory object used to facilitate
	// communication between modules and Wraith.
	sharedMemory SharedMemory

	// A mutex to protect access to SHM_WRAITH_STATUS. While
	// SharedMemory is threadsafe for individual calls, checking
	// and changing SHM_WRAITH_STATUS requires two separate calls.
	statusMutex sync.Mutex

	// An instance of WraithConf storing all configuration necessary
	// for Wraith to work correctly.
	conf Config

	// A list of modules available to Wraith
	modules map[string]Module
}

// Spawn an instance of Wraith running synchronously. If you would
// like Wraith to run asynchronously, start this function in a
// goroutine. It can then be stopped with Wraith.Kill().
//
// The first argument is an instance of WraithConf containing the
// configuration for this instance of Wraith. It should be fully
// initialised and filled out. An uninitialised config can lead to
// undefined behaviour.
//
// The following arguments are modules which should be available to
// Wraith. In case of a name conflict, the first module in the
// list with the name will be chosen, the others will be discarded.
//
// Modules are initialised and started in the order they are given.
// It is highly recommended to pass the comms manager module first
// (possibly preceded by modules it depends on) to make sure module
// communications are not lost.
func (w *Wraith) Spawn(conf Config, modules ...Module) {
	// Make sure only one instance runs
	// If another instance is in any state but inactive, exit immediately
	w.statusMutex.Lock()
	if status := w.sharedMemory.Get(SHM_WRAITH_STATUS); status != WSTATUS_INACTIVE && status != nil {
		w.statusMutex.Unlock()
		return
	}
	w.sharedMemory.Set(SHM_WRAITH_STATUS, WSTATUS_ACTIVE)
	w.statusMutex.Unlock()

	// Take note of start time
	w.initTime = time.Now()

	// Save a copy of the config
	w.conf = conf

	// Watch various special cells in shared memory
	statusWatcher, _ := w.sharedMemory.Watch(SHM_WRAITH_STATUS)
	reloadTrigger, _ := w.sharedMemory.Watch(SHM_RELOAD_TRIGGER)

	// Init map of modules
	w.modules = make(map[string]Module)

	// Prepare on-exit cleanup
	defer func() {
		// Always stop all modules before exiting
		// TODO: Note errors
		for _, module := range w.modules {
			module.Stop()
		}

		// Mark Wraith as dead
		w.statusMutex.Lock()
		w.sharedMemory.Set(SHM_WRAITH_STATUS, WSTATUS_INACTIVE)
		w.statusMutex.Unlock()
	}()

	// Save a copy of the passed modules in the `modules` field, using the
	// module name as the key. Also init and start the modules while we're
	// at it.
	// TODO: Note errors
	for _, module := range modules {
		// Ignore duplicates
		if _, exists := w.modules[module.Name()]; !exists {
			w.modules[module.Name()] = module
			module.WraithModuleInit(w)
			module.Start()
		}
	}

	// Run mainloop
	// This is the place where any functions which need to be
	// carried out by Wraith itself are handled, based on an event
	// loop. Most functions are carried out by modules, so there
	// shouldn't be too much here.
	for {
		select {
		case newStatus := <-statusWatcher:
			if newStatus == WSTATUS_DEACTIVATING {
				// On exit trigger (deactivating status), return from the
				// method. All exit cleanup is deferred so it is guaranteed to run.
				return
			}
		case <-reloadTrigger:
			// On reload trigger, restart all modules.
			// TODO: Note errors
			for _, module := range w.modules {
				module.Stop()
				module.Start()
			}
		}
	}
}

// Stop the Wraith instance including all modules. This will
// block until Wraith exits or the provided timeout is reached.
// If the timeout is reached, the method will return false to
// show that it was unable to confirm Wraith's exit.
func (w *Wraith) Kill(timeout time.Duration) bool {
	w.statusMutex.Lock()

	// Trigger exit of mainloop if it's running, otherwise there's nothing to do
	if status := w.sharedMemory.Get(SHM_WRAITH_STATUS); status == WSTATUS_ACTIVE {
		// Watch the SHM_WRAITH_STATUS cell to catch when Wraith exits, and
		// unwatch it straight after we return
		statusWatch, statusWatchId := w.sharedMemory.Watch(SHM_WRAITH_STATUS)
		defer w.sharedMemory.Unwatch(SHM_WRAITH_STATUS, statusWatchId)

		// Trigger exit
		w.sharedMemory.Set(SHM_WRAITH_STATUS, WSTATUS_DEACTIVATING)

		// Unlock mutex so Wraith can transition into inactive state
		// avoiding deadlock
		w.statusMutex.Unlock()

		// Wait for exit or timeout
		timeoutTimer := time.After(timeout)
		for {
			select {
			case status := <-statusWatch:
				if status == WSTATUS_INACTIVE {
					return true
				}
			case <-timeoutTimer:
				return false
			}
		}
	}

	w.statusMutex.Unlock()

	// Wraith is not running anyway so return true
	return true
}

//
// Proxy Methods
//

// These are methods which allow access to Wraith's internal
// properties in a limitted manner, to make sure all access
// is safe and will not cause unexpected behaviour.

// InitTime

// Return the time at which Wraith begun initialisation (recorded
// as soon as Wraith confirms that it is the only running instance).
// This will be the time.Time zero value if Wraith has not yet
// started initialisation.
func (w *Wraith) GetInitTime() time.Time {
	return w.initTime
}

// Fingerprint

// Return Wraith's fingerprint as generated by the configured
// generator. This method checks if the fingerprint has been
// cached and returns the cached value if so. Otherwise, it
// will run the generator function.
func (w *Wraith) GetFingerprint() string {
	if w.fingerprint == "" {
		w.fingerprint = w.conf.FingerprintGenerator()
	}
	return w.fingerprint
}

// SharedMemory

// Proxy to SharedMemory.Get()
func (w *Wraith) SHMGet(cellname string) interface{} {
	return w.sharedMemory.Get(cellname)
}

// Proxy to SharedMemory.Set()
// Disallows writing to protected cells and returns an error
// if a write to such is attempted.
func (w *Wraith) SHMSet(cellname string, value interface{}) error {
	for _, protectedCell := range []string{
		SHM_WRAITH_STATUS,
	} {
		if cellname == protectedCell {
			return fmt.Errorf("%s is a protected cell", cellname)
		}
	}

	w.sharedMemory.Set(cellname, value)
	return nil
}

// Proxy to SharedMemory.Watch()
func (w *Wraith) SHMWatch(cellname string) (chan interface{}, int) {
	return w.sharedMemory.Watch(cellname)
}

// Proxy to SharedMemory.Unwatch()
func (w *Wraith) SHMUnwatch(cellname string, watchId int) {
	w.sharedMemory.Unwatch(cellname, watchId)
}
