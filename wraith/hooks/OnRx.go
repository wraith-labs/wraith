package hooks

import "sync"

// OnRx hook structure
type onRxHook func([]byte) []string

// List of OnRx hooks structure
type onRxHookList struct {
	data  []onRxHook
	mutex sync.Mutex
}

func (l *onRxHookList) Add(hook onRxHook) {
	defer l.mutex.Unlock()
	l.mutex.Lock()
	l.data = append(l.data, hook)
}
func (l *onRxHookList) Range(f func(hook onRxHook)) {
	for _, hook := range l.data {
		f(hook)
	}
}

// List of hooks applied to the OnRx event
var OnRx onRxHookList

// Trigger function for OnRx event
func RunOnRx(data []byte) []string {
	results := []string{}
	OnRx.Range(func(hook onRxHook) {
		if result := hook(data); len(result) > 0 {
			results = append(results, result...)
		}
	})
	return results
}
