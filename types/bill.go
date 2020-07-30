package types

import (
	"github.com/Qitmeer/qitmeer-wallet/wtxmgr"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/types"
)

type Bill struct {
	TxID   hash.Hash
	Amount types.Amount
	Block  wtxmgr.Block
}
type Bills []Bill

func (s Bills) Len() int { return len(s) }

func (s Bills) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s Bills) Less(i, j int) bool {
	return s[i].Block.Height >= s[j].Block.Height
}
