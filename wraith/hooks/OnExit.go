package hooks

import "sync"

// OnExit hook structure
type onExitHook func()

// List of OnExit hooks structure
type onExitHookList struct {
	data  []onExitHook
	mutex sync.Mutex
}

func (l *onExitHookList) Add(hook onExitHook) {
	defer l.mutex.Unlock()
	l.mutex.Lock()
	l.data = append(l.data, hook)
}
func (l *onExitHookList) Range(f func(hook onExitHook)) {
	for _, hook := range l.data {
		f(hook)
	}
}

// List of hooks applied to the OnExit event
var OnExit onExitHookList

// Trigger function for OnExit event
func RunOnExit() {
	OnExit.Range(func(hook onExitHook) {
		hook()
	})
}
