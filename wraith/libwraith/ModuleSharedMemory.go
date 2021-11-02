package libwraith

import (
	"sync"
	"time"
)

// A struct for storing individual pieces of data within the
// ModuleSharedMemory. Using a struct over simply values in a
// map allows for storing additional metadata and simpler
// interaction with the shared memory (ie. watchers can be
// handled by the cell and don't need to be kept track of by
// the memory).
type moduleSharedMemoryCell struct {
	data           interface{}
	watchers       map[int]chan interface{}
	watcherCounter int
	mutex          sync.Mutex
}

// Lock the mutex for the cell and return the function to unlock
// the mutex. This allows for a simple, one-liner to lock and unlock
// the mutex at the top of every method of the cell like so:
// `defer c.autolock()()`
func (c *moduleSharedMemoryCell) autolock() func() {
	c.mutex.Lock()
	return c.mutex.Unlock
}

// Notify watchers of this cell about the current value of the cell.
// This is a helper which should be called whenever the value is
// changed by one of the other methods. It doesn't lock the cell
// because it assumes the caller already did this. If it did not
// and the value changes while looping, different watchers could
// get different values, which could be bad.
//
// All pushes to channels are done asynchronously as to return as
// quickly as possible and therefore reduce the time taken to set
// cells. This also means that watchers get updates as quickly as
// possible. However, a goroutine is spawned for each watcher because
// of this, though this should be fine because goroutines have
// minimal overhead.
//
// Pushes time out after one second, so if a channel is full for
// longer than that, the watcher which owns that channel will not
// receive that update.
func (c *moduleSharedMemoryCell) notify() {
	const TIMEOUT = time.Second * 1

	for watcherId, watcherChannel := range c.watchers {
		go func() {
			// The channel could be closed, in which case a panic will
			// occur. We don't want any panics so we will catch it here.
			// However, there is no point ever trying to send to this
			// channel again, so it should be removed.
			defer func() {
				if r := recover(); r != nil {
					delete(c.watchers, watcherId)
				}
			}()

			// Send to channel with timeout
			select {
			case watcherChannel <- c.data:
			case <-time.After(TIMEOUT):
			}
		}()
	}
}

// Set the value of the cell to that passed as the argument. This
// will also notify all watchers of the change.
func (c *moduleSharedMemoryCell) Set(value interface{}) {
	defer c.autolock()()

	c.data = value

	c.notify()
}

// Get the current value of the cell.
func (c *moduleSharedMemoryCell) Get() (value interface{}) {
	defer c.autolock()()

	return c.data
}

// Add a channel to the list of watchers for this cell. This means
// that the channel will receive the value of this cell whenever it
// changes. Returns the assigned ID of the channel which can be
// used to unwatch the cell.
func (c *moduleSharedMemoryCell) Watch(channel chan interface{}) int {
	// Defer statements are executed in LIFO order so the counter
	// will be incremented and then the mutex will be unlocked.
	defer c.autolock()()
	defer func() { c.watcherCounter++ }()

	c.watchers[c.watcherCounter] = channel

	return c.watcherCounter
}

// Remove a channel from the list of watchers from this cell. This
// means that the channel will no longer receive updates when the
// value of this cell changes. Takes the ID returned by Watch().
func (c *moduleSharedMemoryCell) Unwatch(id int) {
	defer c.autolock()

	delete(c.watchers, id)
}

// A struct for sharing memory between modules in a flexible way.
// It allows modules to write to memory
type ModuleSharedMemory struct {
	mem map[string]moduleSharedMemoryCell
}

// Create a cell with the given name and return its pointer.
func (m *ModuleSharedMemory) createcell(name string) *moduleSharedMemoryCell {
	cell := moduleSharedMemoryCell{}
	m.mem[name] = cell
	return &cell
}

// Initialise the MSM. This should be called only once as it
// initialises the internal data structure causing all cells
// to be reset if they currently store data. However, it must
// be called before any other methods are called else they will
// panic.
func (m *ModuleSharedMemory) Init() {
	m.mem = make(map[string]moduleSharedMemoryCell)
}

// Set the value of the given cell to that passed as the argument.
// This will also notify all watchers of the change.
func (m *ModuleSharedMemory) Set(cellName string, value interface{}) {
	// If the cell exists...
	if cell, exists := m.mem[cellName]; exists {
		// ...set its value
		cell.Set(value)
	} else {
		// ...create the cell, then set its value
		m.createcell(cellName).Set(value)
	}
}

// Get the current value of a given cell.
func (m *ModuleSharedMemory) Get(cellName string) interface{} {
	// If the cell exists...
	if cell, exists := m.mem[cellName]; exists {
		// ...return its value
		return cell.Get()
	} else {
		// ...return nil because the cell is nil
		return nil
	}
}

// Add a channel to the list of watchers for this cell. This means
// that the channel will receive the value of this cell whenever it
// changes. If the cell does not exist, it will be created as to
// allow watching for cells to be created in the future. Returns
// the channel which will receive updates and the ID assigned to that
// channel which can be used to unwatch the cell.
func (m *ModuleSharedMemory) Watch(cellName string) (channel chan interface{}, watchId int) {
	// Create a channel, to be used for sending updates, with no buffer
	channel = make(chan interface{})

	// If the cell exists...
	if cell, exists := m.mem[cellName]; exists {
		// ...add a watcher
		watchId = cell.Watch(channel)
	} else {
		// ...create the cell, then add a watcher
		watchId = m.createcell(cellName).Watch(channel)
	}

	return channel, watchId
}

// Remove a channel from the list of watchers from a given cell.
// This means that the channel will no longer receive updates
// when the value of this cell changes. Takes the ID returned
// by Watch().
func (m *ModuleSharedMemory) Unwatch(cellName string, watchId int) {
	// If the cell exists...
	if cell, exists := m.mem[cellName]; exists {
		// ...remove the watcher (if the ID doesn't exist, this
		// is a no-op)
		cell.Unwatch(watchId)
	}
	// ...otherwise, there's nothing to do
}
