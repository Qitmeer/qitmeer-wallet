package qitmeerd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Qitmeer/qitmeer-wallet/wallet"

	"github.com/Qitmeer/qitmeer/log"
)

// Qitmeerd mgr
type Qitmeerd struct {
	Status *Status // *qJson.InfoNodeResult
	Wt     *wallet.Wallet
}

// NewQitmeerd make qitmeerd
func NewQitmeerd(wt *wallet.Wallet, name string) *Qitmeerd {
	d := &Qitmeerd{
		Wt:     wt,
		Status: &Status{Network: wt.ChainParams().Name, CurrentName: name},
	}
	d.Start()
	return d
}

// Start run
func (qitmeerd *Qitmeerd) Start() {
	go qitmeerd.GetStatus()
}

// GetStatus get current qitmeerd status
func (qitmeerd *Qitmeerd) GetStatus() {
	defer func() {
		if rev := recover(); rev != nil {
			go qitmeerd.GetStatus()
			log.Error("qitmeerd GetStatus recover", "recover", rev)
		}
	}()

	ticker := time.NewTicker(30 * time.Second)
	for {
		select {
		case <-ticker.C:
			if qitmeerd.Wt.HttpClient == nil {
				log.Debug("qitmeerd GetNodeInfo,but HttpClient nil")
				continue
			}

			nodeInfo, err := qitmeerd.Wt.HttpClient.GetNodeInfo()
			if err != nil {
				qitmeerd.Status.err = fmt.Sprintf("getNodeInfo err: %v", err)
				log.Error("qitmeerd GetNodeInfo err", "err", err)
				continue
			}

			qitmeerd.Status.err = ""
			qitmeerd.Status.MainOrder = nodeInfo.GraphState.MainOrder
			qitmeerd.Status.MainHeight = nodeInfo.GraphState.MainHeight
			qitmeerd.Status.Blake2bdDiff = strconv.FormatFloat(nodeInfo.PowDiff.Blake2bdDiff, 'f', 2, 64)
			qitmeerd.Status.CuckarooDiff = nodeInfo.PowDiff.CuckarooDiff
			qitmeerd.Status.CuckatooDiff = nodeInfo.PowDiff.CuckatooDiff
		}
	}
}
