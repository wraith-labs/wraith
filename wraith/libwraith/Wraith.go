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

// Spawn an instance of Wraith running synchronously. If you would
// like Wraith to run asynchronously, start this function in a
// goroutine.
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
func (w *Wraith) Spawn(conf WraithConf, modules ...WraithModule) {
	// Take note of start time
	w.initTime = time.Now()

	// Init dead channel to signal Wraith's status
	w.IsDead = make(chan struct{})

	// Init shared memory so it's usable, as it is needed
	// throughout the following.
	w.SharedMemory.Init()

	// Watch various special cells in shared memory
	exitTrigger, _ := w.SharedMemory.Watch(SHM_EXIT_TRIGGER)
	reloadTrigger, _ := w.SharedMemory.Watch(SHM_RELOAD_TRIGGER)

	// Prepare on-exit cleanup
	defer func() {
		// Always stop all modules before exiting
		// TODO: Note errors
		for _, module := range w.Modules {
			module.Stop()
		}

		// Mark Wraith as dead by closing dead channel
		close(w.IsDead)
	}()

	// Save a copy of the passed modules in the `modules` field, using the
	// module name as the key. Also init and start the modules while we're
	// at it.
	// TODO: Note errors
	for _, module := range modules {
		// Ignore duplicates
		if _, exists := w.Modules[module.Name()]; !exists {
			w.Modules[module.Name()] = module
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
		case <-exitTrigger:
			// On exit trigger, return from the method. All exit cleanup
			// is deferred so it is guaranteed to run.
			return
		case <-reloadTrigger:
			// On reload trigger, restart all modules.
			// TODO: Note errors
			for _, module := range w.Modules {
				module.Stop()
				module.Start()
			}
		}
	}
}

// Stop the Wraith instance including all modules. This will
// block until Wraith exits.
func (w *Wraith) Kill() {
	// Trigger exit of mainloop
	w.SharedMemory.Set(SHM_EXIT_TRIGGER, true)

	// Await death
	<-w.IsDead
}
