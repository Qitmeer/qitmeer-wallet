package types

import (
	"github.com/Qitmeer/qitmeer/common/hash"
)

//  effects that a transaction makes on a specific address
type Payment struct {
	TxID      hash.Hash
	Variation int64	// deposit: >0, withdraw: <= 0

	BlockHash  hash.Hash
	BlockOrder uint32
}

//  log of payments
type Bill []Payment

func (b Bill) Len() int { return len(b) }

func (b Bill) Swap(i, j int) { b[i], b[j] = b[j], b[i] }

func (b Bill) Less(i, j int) bool {
	return b[i].BlockOrder >= b[j].BlockOrder
}
