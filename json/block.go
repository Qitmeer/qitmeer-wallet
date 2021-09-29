package json

import (
	"github.com/Qitmeer/qitmeer/core/json"
	"time"
)

type BlockHttpResult struct {
	Hash          string `json:"hash"`
	Confirmations uint32 `json:"confirmations"`
	Version       int32  `json:"version"`
	Weight       int64              `json:"weight"`
	Order        uint32             `json:"order"`
	TxRoot       string             `json:"txRoot"`
	Transactions []json.TxRawResult `json:"transactions,omitempty"`
	StateRoot    string             `json:"stateRoot"`
	Bits         string             `json:"bits"`
	Difficulty   float64            `json:"difficulty"`
	Nonce        uint64             `json:"nonce"`
	Timestamp    time.Time          `json:"timestamp"`
	Parents      []string           `json:"parents"`
	Children     []string           `json:"children"`
	Txsvalid bool `json:"txsvalid"`
	IsBlue   bool `json:"isblue"`
}
type PageTxRawResult struct {
	Total        int32              `json:"total"`
	Page         int32              `json:"page"`
	PageSize     int32              `json:"page_size"`
	Transactions []json.TxRawResult `json:"transactions,omitempty"`
}
