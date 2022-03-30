package libwraith

import (
	"sync"
	"time"
)

// A struct for storing individual pieces of data within the
// SharedMemory. Using a struct over simply values in a
// map allows for storing additional metadata and simpler
// interaction with the shared memory (ie. watchers can be
// handled by the cell and don't need to be kept track of by
// the memory).
type shmCell struct {
	data           any
	watchers       map[int]chan any
	watcherCounter int
}

// Initialise the cell so that it's useable. Calling the cell's other
// methods before this one can lead to panics. This should be called
// exactly once as each consecutive call effectively wipes the cell.
func (c *shmCell) init() {
	c.watchers = make(map[int]chan any)
	c.watcherCounter = 0
}

// Notify watchers of this cell about the current value of the cell.
// This is a helper which should be called whenever the value is
// changed by one of the other methods.
//
// All pushes to channels are done asynchronously as to return as
// quickly as possible and therefore reduce the time taken to set
// cells. This also means that watchers get updates as quickly as
// possible. However, a goroutine is spawned for each watcher because
// of this, though this should be fine because goroutines have
// minimal overhead. The call will block until all goroutines return.
//
// Pushes time out after SHARED_MEMORY_WATCHER_NOTIF_TIMEOUT seconds,
// so if a channel is full for longer than that, the watcher which
// owns that channel will not receive that update.
func (c *shmCell) notify() {
	wg := sync.WaitGroup{}
	wg.Add(len(c.watchers))

	for watcherId, watcherChannel := range c.watchers {
		go func(watcherId int, watcherChannel chan any) {
			// At the very end, mark this goroutine as done
			defer wg.Done()
			// The channel could be closed, in which case a panic will
			// occur. We don't want any panics so we will catch it here.
			// However, there is no point ever trying to send to this
			// channel again, so it should be removed.
			// TODO: Not relying on panics would be nice
			defer func() {
				if r := recover(); r != nil {
					delete(c.watchers, watcherId)
				}
			}()

			// Send to channel with timeout
			select {
			case watcherChannel <- c.data:
			case <-time.After(SHMCONF_WATCHER_NOTIF_TIMEOUT * time.Second):
			}
		}(watcherId, watcherChannel)
	}

	// Wait for all goroutines to finish, otherwise this function would
	// return, the SharedMemory might release the lock, another call might be
	// made to change the value and different watchers would get different
	// updates. As the goroutines have timeouts, this shouldn't take very
	// long.
	wg.Wait()
}

// Set the value of the cell to that passed as the argument. This
// will also notify all watchers of the change.
func (c *shmCell) set(value any) {
	c.data = value

	c.notify()
}

// Get the current value of the cell.
func (c *shmCell) get() (value any) {
	return c.data
}

// Add a channel to the list of watchers for this cell. This means
// that the channel will receive the value of this cell whenever it
// changes. Returns the assigned ID of the channel which can be
// used to unwatch the cell.
func (c *shmCell) watch(channel chan any) int {
	defer func() { c.watcherCounter++ }()

	c.watchers[c.watcherCounter] = channel

	return c.watcherCounter
}

// Remove a channel from the list of watchers from this cell. This
// means that the channel will no longer receive updates when the
// value of this cell changes. Takes the ID returned by Watch().
func (c *shmCell) unwatch(id int) {
	delete(c.watchers, id)
}

// A struct for sharing memory between modules and Wraith in a
// thread-safe way while providing facilities to watch individual
// memory cells for updates.
type shm struct {
	isPostInit bool

	// This is an RWMutex because eventually it could be used to
	// improve performance. Currently even reading requires a write
	// lock as the shm is autoinitialised which is a write operation.
	// TODO
	mutex sync.RWMutex

	mem map[string]*shmCell
}

// Initialise the SM if it's not already initialised. This requires
// a lock, but assumes that this is handled by the caller.
func (m *shm) initIfNot() {
	if !m.isPostInit {
		m.mem = make(map[string]*shmCell)
		m.isPostInit = true
	}
}

// Lock the mutex and return the function to unlock it. This
// allows for a simple, one-liner to lock and unlock the mutex
// at the top of every method like so: `defer m.autolock()()`.
func (m *shm) autolock() func() {
	m.mutex.Lock()
	return m.mutex.Unlock
}

// Same as autolock but locks for reading only
func (m *shm) rautolock() func() {
	m.mutex.RLock()
	return m.mutex.RUnlock
}

// Create and init a cell with the given name and return its pointer.
func (m *shm) createcell(name string) *shmCell {
	m.mem[name] = &shmCell{}
	m.mem[name].init()
	return m.mem[name]
}

// Set the value of the given cell to that passed as the argument.
// This will also notify all watchers of the change.
func (m *shm) Set(cellName string, value any) {
	defer m.autolock()()
	m.initIfNot()

	// If the cell exists...
	if cell, exists := m.mem[cellName]; exists {
		// ...set its value
		cell.set(value)
	} else {
		// ...create the cell, then set its value
		m.createcell(cellName).set(value)
	}
}

// Get the current value of a given cell.
func (m *shm) Get(cellName string) any {
	defer m.autolock()()
	m.initIfNot()

	// If the cell exists...
	if cell, exists := m.mem[cellName]; exists {
		// ...return its value
		return cell.get()
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
func (m *shm) Watch(cellName string) (channel chan any, watchId int) {
	defer m.autolock()()
	m.initIfNot()

	// Create a channel, to be used for sending updates, with no buffer
	channel = make(chan any, SHMCONF_WATCHER_CHAN_SIZE)

	// If the cell exists...
	if cell, exists := m.mem[cellName]; exists {
		// ...add a watcher
		watchId = cell.watch(channel)
	} else {
		// ...create the cell, then add a watcher
		watchId = m.createcell(cellName).watch(channel)
	}

	return channel, watchId
}

// Remove a channel from the list of watchers from a given cell.
// This means that the channel will no longer receive updates
// when the value of this cell changes. Takes the ID returned
// by Watch().
func (m *shm) Unwatch(cellName string, watchId int) {
	defer m.autolock()()
	m.initIfNot()

	// If the cell exists...
	if cell, exists := m.mem[cellName]; exists {
		// ...remove the watcher (if the ID doesn't exist, this
		// is a no-op)
		cell.unwatch(watchId)
	}
	// ...otherwise, there's nothing to do
}
