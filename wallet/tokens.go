package wallet

import (
	json2 "encoding/json"
	"fmt"
	"github.com/Qitmeer/qitmeer-wallet/walletdb"
	"github.com/Qitmeer/qng/core/json"
	"github.com/Qitmeer/qng/core/types"
	"sync"
)

type QitmeerToken struct {
	tokens map[string]*json.TokenState
	lock   sync.RWMutex
}

func NewQitmeerToken(ns walletdb.ReadWriteBucket) *QitmeerToken {
	tokens := make(map[string]*json.TokenState, 0)
	_ = ns.ForEach(func(k, v []byte) error {
		token, err := DecodeToken(v)
		if err == nil {
			tokens[string(k)] = token
			types.CoinNameMap[types.CoinID(token.CoinId)] = token.CoinName
			return nil
		} else {
			return err
		}
	})

	return &QitmeerToken{
		tokens: tokens,
		lock:   sync.RWMutex{},
	}
}

func (q *QitmeerToken) Add(t json.TokenState) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.tokens[t.CoinName] = &t
}

func (q *QitmeerToken) GetToken(coin string) (*json.TokenState, error) {
	q.lock.RLock()
	defer q.lock.RUnlock()

	token, ok := q.tokens[coin]
	if ok {
		return token, nil
	}
	return nil, fmt.Errorf("coin %s dose not exist", coin)
}

func (q *QitmeerToken) Encode() []byte {
	bytes, _ := json2.Marshal(q)
	return bytes
}

func EncodeToken(state json.TokenState) []byte {
	bytes, _ := json2.Marshal(state)
	return bytes
}

func DecodeToken(bytes []byte) (*json.TokenState, error) {
	var t = &json.TokenState{}
	err := json2.Unmarshal(bytes, t)
	return t, err
}
