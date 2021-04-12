package parts

import (
	"fmt"
	"os"
	"os/user"
	"regexp"
	"runtime"

	"github.com/0x1a8510f2/wraith/config"
	"github.com/0x1a8510f2/wraith/proto"
)

func init() {
	proto.PartMap.Add("w.validity", func(hkvs *proto.HandlerKeyValueStore, data interface{}) {
		isValid := false

		defer func() {
			if isValid {
				hkvs.Set("w.validity.isValid", true)
			} else {
				hkvs.Set("w.validity.isValid", false)
			}
		}()

		// Enforce validity constraints

		// If there are validity constraints (this function was called), but they are incorrectly formatted,
		// always assume invalid
		if validity, ok := data.(map[string]string); ok {
			// Wraith Fingerprint/ID restriction
			if constraint, ok := validity["wfpt"]; ok {
				// Always fail if an ID restriction is present and Wraith has not been given an ID
				if config.Config.Wraith.Fingerprint == "" {
					return
				}
				match, err := regexp.Match(constraint, []byte(config.Config.Wraith.Fingerprint))
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
	})
}
