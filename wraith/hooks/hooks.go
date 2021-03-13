package hooks

// Hook types
type onStartHook func()
type onCommandHook func(string) string
type onExitHook func()

// Hook lists
var OnStart []onStartHook
var OnCommand []onCommandHook
var OnExit []onExitHook

// Hook runners
func RunOnStart() {
	for _, hook := range OnStart {
		hook()
	}
}

func RunOnCommand(command string) []string {
	results := []string{}
	for _, hook := range OnCommand {
		results = append(results, hook(command))
	}
	return results
}

func RunOnExit() {
	for _, hook := range OnExit {
		hook()
	}
}
