package stdmod

import (
	"context"
	"fmt"
	"sync"
	"time"

	"git.0x1a8510f2.space/0x1a8510f2/wraith/wraith/libwraith"
)

// A basic debug module which logs some events and attempts to use some
// features of Wraith to make sure they work.
type DebugModule struct {
	mutex sync.Mutex
}

func (m *DebugModule) Mainloop(ctx context.Context, w *libwraith.Wraith) {
	fmt.Printf("Starting the debug module!\n")

	fmt.Printf("Wraith strain ID: %s\nWraith fingerprint: %s\n", w.GetStrainId(), w.GetFingerprint())

	// Ensure this instance is only started once and mark as running if so
	single := m.mutex.TryLock()
	if !single {
		fmt.Printf("It seems this debug module is already running! Exiting...\n")
		panic(fmt.Errorf("already running"))
	}
	defer func() {
		fmt.Printf("Marking debug module as not running!\n")
		m.mutex.Unlock()
	}()

	fmt.Printf("No other debug module instances are running. Good to start!\n")

	// Watch some debug cells
	fmt.Printf("Setting up watch for memory cells!\n")
	debugCellWatch, debugCellWatchId := w.SHMWatch("w.debug")
	errorCellWatch, errorCellWatchId := w.SHMWatch(libwraith.SHM_ERRS)

	// Always cleanup SHM when exiting
	defer func() {
		// Unwatch cells
		fmt.Printf("Cancelling watch for memory cells!\n")
		w.SHMUnwatch("w.debug", debugCellWatchId)
		w.SHMUnwatch(libwraith.SHM_ERRS, errorCellWatchId)
	}()

	// Send something to the debug memory cell
	fmt.Printf("Sending `debugging!` to `w.debug` memory cell!\n")
	w.SHMSet("w.debug", "debugging!")

	// Mainloop
	fmt.Printf("Starting mainloop!\n")
	for {
		select {
		// Trigger exit when requested
		case <-ctx.Done():
			fmt.Printf("Debug module exit requested via context close!\n")
			return
		// Manage w.debug watch
		case value := <-debugCellWatch:
			fmt.Printf("Received value in the `w.debug` memory cell: `%v`\n", value)
		case err := <-errorCellWatch:
			fmt.Printf("Received value in the `%s` memory cell: `%v`\n", libwraith.SHM_ERRS, err)
		case <-time.After(5 * time.Second):
			fmt.Printf("Sending `hello debug!` into the `w.debug` memory cell!\n")
			w.SHMSet("w.debug", "hello debug!")

			fmt.Printf("Checking Wraith alive status!\n")
			alive := w.IsAlive()
			fmt.Printf("Wraith reports as alive: %v\n", alive)
		}
	}
}

// Return the name of this module
func (m *DebugModule) WraithModuleName() string {
	fmt.Printf("Debug module `WraithModuleName()` called!\n")
	return "w.debug"
}
