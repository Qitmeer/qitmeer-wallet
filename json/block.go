package json

import "github.com/HalalChain/qitmeer-lib/core/json"

type BlockHttpResult struct {
	Hash          string        `json:"hash"`
	Confirmations int64         `json:"confirmations"`
	Version       int32         `json:"version"`
	Weight       int64         `json:"weight"`
	Order       int32         `json:"order"`
	TxRoot        string        `json:"txRoot"`
	Transactions  []json.TxRawResult `json:"transactions,omitempty"`
	StateRoot     string         `json:"stateRoot"`
	Bits          string        `json:"bits"`
	Difficulty    float64       `json:"difficulty"`
	Nonce         uint64        `json:"nonce"`
	Timestamp     string     `json:"timestamp"`
	Parents       []string     `json:"parents"`
	Children       []string     `json:"children"`
}
type PageTxRawResult struct {
	Total int32
	Page  int32
	PageSize int32
	Transactions  []json.TxRawResult `json:"transactions,omitempty"`
}
