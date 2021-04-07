package parts

import "github.com/0x1a8510f2/wraith/proto"

func init() {
	proto.PartMap.Add("w.validity", func(proto.HandlerKeyValueStore, interface{}) {})
}
