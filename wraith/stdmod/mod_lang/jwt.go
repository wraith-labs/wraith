package mod_lang

import (
	"fmt"
	"time"

	"git.0x1a8510f2.space/0x1a8510f2/wraith/libwraith"
	"github.com/pascaldekloe/jwt"
)

type JWTModule struct {
	EncodeKey []byte
	DecodeKey []byte
}

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
