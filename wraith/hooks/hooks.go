package hooks

import "sync"

// Hook types
type onStartHook func()
type onRxHook func(map[string]interface{}) string
type onExitHook func()

// Hook list types
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

// Hook lists
var OnStart onStartHookList
var OnRx onRxHookList
var OnExit onExitHookList

// Hook runners
func RunOnStart() {
	OnStart.Range(func(hook onStartHook) {
		hook()
	})
}

func RunOnRx(data map[string]interface{}) []string {
	results := []string{}
	OnRx.Range(func(hook onRxHook) {
		if result := hook(data); result != "" {
			results = append(results, result)
		}
	})
	return results
}

func RunOnExit() {
	OnExit.Range(func(hook onExitHook) {
		hook()
	})
}
