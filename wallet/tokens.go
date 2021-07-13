package wallet

import (
	"fmt"
	"github.com/Qitmeer/qitmeer/core/json"
	"sync"
)

type QitmeerToken struct {
	tokens map[string]*json.TokenState
	lock   sync.RWMutex
}

func NewQitmeerToken() *QitmeerToken {
	return &QitmeerToken{
		tokens: make(map[string]*json.TokenState, 0),
		lock:   sync.RWMutex{},
	}
}

func (q *QitmeerToken) Add(t *json.TokenState) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.tokens[t.CoinName] = t
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
