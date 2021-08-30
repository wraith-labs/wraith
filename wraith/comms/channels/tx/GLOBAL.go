package tx

import "git.0x1a8510f2.space/0x1a8510f2/wraith/types"

var UnifiedTxQueue types.TxQueue

func init() {
	UnifiedTxQueue = make(types.TxQueue)
}
