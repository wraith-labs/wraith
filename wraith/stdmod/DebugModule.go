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
	running      bool
	runningMutex sync.Mutex
}

func (m *DebugModule) Mainloop(ctx context.Context, w *libwraith.Wraith) error {
	// Ensure this instance is only started once and mark as running if so
	fmt.Printf("Starting the debug module!\n")
	m.runningMutex.Lock()
	if m.running {
		m.runningMutex.Unlock()
		fmt.Printf("It seems this debug module is already running! Exiting...\n")
		return fmt.Errorf("already running")
	}
	m.running = true
	m.runningMutex.Unlock()

	fmt.Printf("No other debug module instances are running. Good to start!\n")

	// Always clear running status when exiting
	defer func() {
		fmt.Printf("Marking debug module as not running!\n")
		// Mark as not running internally
		m.runningMutex.Lock()
		m.running = false
		m.runningMutex.Unlock()
	}()

	// Watch some debug cells
	fmt.Printf("Setting up watch for `w.debug` memory cell!\n")
	debugCellWatch, debugCellWatchId := w.SHMWatch("w.debug")

	// Always cleanup SHM when exiting
	defer func() {
		// Unwatch cells
		fmt.Printf("Cancelling watch for `w.debug` memory cell!\n")
		w.SHMUnwatch("w.debug", debugCellWatchId)
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
			return nil
		// Manage w.debug watch
		case value := <-debugCellWatch:
			fmt.Printf("Received value in the `w.debug` memory cell: `%v`\n", value)
		case <-time.After(5 * time.Second):
			fmt.Printf("Sending `hello debug!` into the `w.debug` memory cell!\n")
			w.SHMSet("w.debug", "hello debug!")
		}
	}
}

// Return the name of this module
func (m *DebugModule) WraithModuleName() string {
	fmt.Printf("Debug module `WraithModuleName()` called!\n")
	return "w.debug"
}
