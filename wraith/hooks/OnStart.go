package hooks

import "sync"

// OnStart hook structure
type onStartHook func()

// List of OnStart hooks structure
type onStartHookList struct {
	data  []onStartHook
	mutex sync.Mutex
}

func (l *onStartHookList) Add(hook onStartHook) {
	defer l.mutex.Unlock()
	l.mutex.Lock()
	l.data = append(l.data, hook)
}
func (l *onStartHookList) Range(f func(hook onStartHook)) {
	for _, hook := range l.data {
		f(hook)
	}
}

// List of hooks applied to the OnStart event
var OnStart onStartHookList

// Trigger function for OnStart event
func RunOnStart() {
	OnStart.Range(func(hook onStartHook) {
		hook()
	})
}
