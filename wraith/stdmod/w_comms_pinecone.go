package stdmod

import (
	"crypto/ed25519"

	"git.0x1a8510f2.space/0x1a8510f2/wraith/libwraith"
)

type WCommsPinecone struct {
	wraith *libwraith.Wraith
	pubkey ed25519.PublicKey

	// Configuration properties
	Privkey ed25519.PrivateKey
}

func (wcp *WCommsPinecone) WraithModuleInit(w *libwraith.Wraith) {
	wcp.wraith = w
}

func (wcp *WCommsPinecone) Start() error {
	// Generate the public key from the given private key
	// We could just accept both, but we don't want to risk them not matching
	/*pubkey, ok = wcp.Privkey.Public().(ed25519.PublicKey)
	if !ok {
		// Well, this is awkward, the public key of
	}*/
	return nil
}

func (wcp *WCommsPinecone) Stop() error { return nil }

func (wcp *WCommsPinecone) Name() string { return "w.comms.pinecone" }
