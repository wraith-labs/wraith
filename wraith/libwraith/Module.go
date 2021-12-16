package libwraith

// An interface describing the structure of a Wraith Module
type Module interface {
	// Initialise the module. This is called as soon as Wraith is made aware
	// of a module and will only ever be called once on any instance of
	// the module.
	//
	// The argument is a pointer to the Wraith struct which owns the module.
	// This is used for communication purposes and to allow the module to
	// control Wraith, so it should be saved for the lifetime of the module
	// if needed.
	//
	// This method is called synchronously and will block further execution,
	// so long-running tasks should be started as goroutines.
	//
	// The module should *NOT* run its mainloop here (if any) as this is
	// handled by the `Start()` method, which runs asynchronously.
	//
	// If anything within this method has the potential to error and may need
	// handling, it likely shouldn't be in this method, but in the `Start()`
	// method instead. This method's primary purpose is to initialise
	// properties and save the Wraith pointer, which shouldn't cause errors.
	WraithModuleInit(*Wraith)

	// Start the module's mainloop. This is called when Wraith decides the
	// module should be running and can be called multiple times. Usually,
	// this is after the module is already stopped, but this is not guaranteed
	// to be the case so the method should be robust against multiple calls
	// and only start a single instance.
	//
	// The method runs synchronously so it should exit as soon as possible and
	// start the mainloop in a goroutine.
	//
	// Any errors should ideally be handled within the method and not
	// propagate up the stack. As modules can be very diverse, Wraith
	// is unable to correctly handle module start errors and will resort
	// to taking note of them and moving on.
	Start() error

	// Stop the module's mainloop. This is called in various situations
	// including when Wraith or the host is shutting down, or if the module
	// is deemed unnecessary for the time being.
	//
	// The method runs synchronously so it should exit as soon as it is
	// certain the mainloop has exited.
	//
	// Any errors should ideally be handled within the method and not
	// propagate up the stack. As modules can be very diverse, Wraith
	// is unable to correctly handle module start errors and will resort
	// to taking note of them and moving on.
	//
	// Modules must comply with the stop request as not to cause unexpected
	// behaviour. In the case of host or Wraith shutdown, the main Wraith
	// process may begin teardown, causing the module to fail in unexpected
	// ways.
	Stop() error

	// Return a string representing the name of the module. This is used to
	// generate a map of modules and their names to allow for easy listing,
	// enabling and disabling of modules.
	//
	// The method should be a single-line return statement with a hard-coded
	// string.
	//
	// Module names should be globally unique. Multiple modules using the same
	// name will clash and only one of them will actually be activated.
	// Because of this, module name namespacing is highly recommended. For
	// example, the name "keylogger" is bad, because it's likely to be used by
	// multiple modules. Instead, "io.github.user.keylogger" could be used.
	//
	// Official modules use the `w` namespace.
	Name() string
}
