package hooks

import "sync"

// Hook types
type onStartHook func()
type onCommandHook func(string) string
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

type onCommandHookList struct {
	data  []onCommandHook
	mutex sync.Mutex
}

func (l *onCommandHookList) Add(hook onCommandHook) {
	defer l.mutex.Unlock()
	l.mutex.Lock()
	l.data = append(l.data, hook)
}
func (l *onCommandHookList) Range(f func(hook onCommandHook)) {
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
var OnCommand onCommandHookList
var OnExit onExitHookList

// Hook runners
func RunOnStart() {
	OnStart.Range(func(hook onStartHook) {
		hook()
	})
}

func RunOnCommand(command string) []string {
	results := []string{}
	OnCommand.Range(func(hook onCommandHook) {
		results = append(results, hook(command))
	})
	return results
}

func RunOnExit() {
	OnExit.Range(func(hook onExitHook) {
		hook()
	})
}
