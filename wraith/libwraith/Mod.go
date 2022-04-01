package libwraith

import "context"

// An interface describing the structure of a Wraith Module
type mod interface {
	// Start the module's mainloop. This is called as soon as the module is added to
	// the Wraith and is guaranteed to be called once at a time (that is, it will not
	// be called again until it returns).
	//
	// The method is called asynchronously and should block indefinitely (never return)
	// unless its context is cancelled. If this method returns or panics and the context
	// is not cancelled, it will be assumed to have crashed and will be restarted
	// immediately unless the max configured crashes occur within a configured time
	// at which point it will no longer be restarted for the entire time Wraith is running.
	//
	// The method receives 2 arguments: a context which, when cancelled, should
	// cause the mainloop to exit (return); and a pointer to the module's parent
	// Wraith instance for communication purposes.
	//
	// Any errors should ideally be handled within the method and not propagate
	// up the stack; however, if an error cannot be handled, it should be returned.
	// It may be worth noting though that, as modules can be very diverse, Wraith
	// is unable to correctly handle module errors and will resort to taking note
	// of them (for possible sending to C2 later) and moving on.
	Mainloop(context.Context, *Wraith) error

	// Return a string representing the name of the module. This is used to
	// generate a map of module names to allow for easy listing, and management
	// of modules.
	//
	// The method should consist of only a single return statement with a
	// hard-coded string.
	//
	// Module names should be globally unique. Multiple modules using the same
	// name will clash and only one of them will actually be activated.
	// Because of this, module name namespacing is highly recommended. For
	// example, the name "keylogger" is bad, because it's likely to be used by
	// multiple modules. Instead, "io.github.user.keylogger" could be used.
	//
	// Official modules use the special `w` namespace. Unofficial modules MUST NOT
	// use this namespace.
	WraithModuleName() string
}
