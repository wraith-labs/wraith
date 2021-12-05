package libwraith

import (
	"time"
)

type Wraith struct {
	// Internal Information

	initTime time.Time

	// Other

	Conf         WraithConf
	SharedMemory SharedMemory
	Modules      map[string]WraithModule
	IsDead       chan struct{}
}

// Initialise and start all stored modules in one go
// This is more efficient than doing both separately
func (w *Wraith) initstartmodules() []error {
	errs := []error{}
	for _, module := range w.Modules {
		module.WraithModuleInit(w)
		err := module.Start()
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// Init all stored modules
/*func (w *Wraith) initmodules() {
	for _, module := range w.Modules {
		module.WraithModuleInit(w)
	}
}*/

// Start all stored modules
/*func (w *Wraith) startmodules() []error {
	errs := []error{}
	for _, module := range w.Modules {
		err := module.Start()
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}*/

// Stop all stored modules
func (w *Wraith) stopmodules() []error {
	errs := []error{}
	for _, module := range w.Modules {
		err := module.Stop()
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// Restart all stored modules
// This is more efficient than stopping and starting
// separately
func (w *Wraith) restartmodules() []error {
	errs := []error{}
	for _, module := range w.Modules {
		err := module.Start()
		if err != nil {
			errs = append(errs, err)
		}
		err2 := module.Stop()
		if err2 != nil {
			errs = append(errs, err2)
		}
	}
	return errs
}

// Spawn an instance of Wraith running synchronously
func (w *Wraith) Spawn(conf WraithConf, modules ...WraithModule) {
	// Take note of start time
	w.initTime = time.Now()

	// Init dead channel to signal Wraith's status
	w.IsDead = make(chan struct{})

	// Init shared memory so it's useable, as it is needed
	// throughout the following.
	w.SharedMemory.Init()

	// Watch various special cells in shared memory
	exitTrigger, _ := w.SharedMemory.Watch("w.exitTrigger")
	reloadTrigger, _ := w.SharedMemory.Watch("w.reloadTrigger")

	// Save a copy of the passed modules in the `modules` field, using the
	// module name as the key
	for _, module := range modules {
		w.Modules[module.Name()] = module
	}

	// Prepare on-exit cleanup
	defer func() {
		// Always stop all modules before exiting
		// TODO: Note errors
		_ = w.stopmodules()

		// Mark Wraith as dead by closing dead channel
		close(w.IsDead)
	}()

	// Init and start all provided modules
	// TODO: Note errors
	_ = w.initstartmodules()

	// Run mainloop
	// This is the place where any functions which need to be
	// carried out by Wraith itself are handled, based on an event
	// loop. Most functions are carried out by modules, so there
	// shouldn't be too much here.
	for {
		select {
		case <-exitTrigger:
			// On exit trigger, return from the method. All exit cleanup
			// is deferred so it is guaranteed to run.
			return
		case <-reloadTrigger:
			// On reload trigger, restart all modules.
			w.restartmodules()
		}
	}
}

// Stop the Wraith instance including all modules. This will
// block until Wraith exits.
func (w *Wraith) Kill() {
	// Trigger exit of mainloop
	w.SharedMemory.Set("w.exitTrigger", true)

	// Await death
	<-w.IsDead
}
