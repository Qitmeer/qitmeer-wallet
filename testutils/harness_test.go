// Copyright (c) 2020 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package testutils

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Qitmeer/qng/core/types"
	"github.com/Qitmeer/qng/params"
)

func TestHarness(t *testing.T) {
	args := []string{"--modules=miner", "--modules=qitmeer", "--notls"}
	h, err := NewHarness(t, params.PrivNetParam.Params, args...)
	if err != nil {
		t.Errorf("create new test harness instance failed %v", err)
		return
	}
	if err := h.Setup(); err != nil {
		t.Errorf("setup test harness instance failed %v", err)
	}

	h2, err := NewHarness(t, params.PrivNetParam.Params, args...)
	defer func() {

		if err := h.Teardown(); err != nil {
			t.Errorf("tear down test harness instance failed %v", err)
		}
		numOfHarnessInstances := len(AllHarnesses())
		if numOfHarnessInstances != 10 {
			t.Errorf("harness num is wrong, expect %d , but got %d", 10, numOfHarnessInstances)
			for _, h := range AllHarnesses() {
				t.Errorf("%v\n", h.Id())
			}
		}

		if err := TearDownAll(); err != nil {
			t.Errorf("tear down all error %v", err)
		}
		numOfHarnessInstances = len(AllHarnesses())
		if numOfHarnessInstances != 0 {
			t.Errorf("harness num is wrong, expect %d , but got %d", 0, numOfHarnessInstances)
			for _, h := range AllHarnesses() {
				t.Errorf("%v\n", h.Id())
			}
		}

	}()
	numOfHarnessInstances := len(AllHarnesses())
	if numOfHarnessInstances != 2 {
		t.Errorf("harness num is wrong, expect %d , but got %d", 2, numOfHarnessInstances)
	}
	if err := h2.Teardown(); err != nil {
		t.Errorf("teardown h2 error:%v", err)
	}

	numOfHarnessInstances = len(AllHarnesses())
	if numOfHarnessInstances != 1 {
		t.Errorf("harness num is wrong, expect %d , but got %d", 1, numOfHarnessInstances)
	}
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			args := []string{"--modules=miner", "--modules=qitmeer", "--notls"}
			NewHarness(t, params.PrivNetParam.Params, args...)
			wg.Done()
		}()
	}

	wg.Wait()
}

func TestSyncUnConfirmedCoinBase(t *testing.T) {
	args := []string{"--modules=miner", "--modules=qitmeer", "--notls"}
	h, err := NewHarness(t, params.PrivNetParam.Params, args...)
	defer h.Teardown()

	if err != nil {
		t.Errorf("new harness failed: %v", err)
		return
	}
	err = h.Setup()
	if err != nil {
		t.Errorf("setup harness failed:%v", err)
		return
	}

	h.WaitWalletInit()
	if info, err := h.Client.NodeInfo(); err != nil {
		t.Errorf("test failed : %v", err)
		return
	} else {
		expect := "privnet"
		if info.Network != expect {
			t.Errorf("test failed, expect %v , but got %v", expect, info.Network)
			return
		}
	}
	GenerateBlock(t, h, 10)
	time.Sleep(10 * time.Second)
	b, err := h.wallet.Balance(types.MEERA)
	if err != nil {
		t.Errorf("test failed:%v", err)
		return
	}
	if b.UnconfirmedAmount.Value != 500000000000 {
		t.Errorf("test failed, expect balance %d, but got %d", 500000000000, b.UnconfirmedAmount.Value)
		return
	}
}

func TestSyncConfirmedCoinBase(t *testing.T) {
	args := []string{"--modules=miner", "--modules=qitmeer", "--notls"}
	h, err := NewHarness(t, params.PrivNetParam.Params, args...)
	defer h.Teardown()

	if err != nil {
		t.Errorf("new harness failed: %v", err)
		return
	}
	err = h.Setup()
	if err != nil {
		t.Errorf("setup harness failed:%v", err)
		return
	}

	h.WaitWalletInit()
	if info, err := h.Client.NodeInfo(); err != nil {
		t.Errorf("test failed : %v", err)
		return
	} else {
		expect := "privnet"
		if info.Network != expect {
			t.Errorf("test failed, expect %v , but got %v", expect, info.Network)
			return
		}

	}
	GenerateBlock(t, h, 18)
	time.Sleep(10 * time.Second)
	b, err := h.wallet.Balance(types.MEERA)
	if err != nil {
		t.Errorf("test failed : %v", err)
		return
	}
	if b.UnspentAmount.Value != 100000000000 {
		t.Errorf("test failed, expect unspent balance %d, but got %d", 100000000000, b.UnspentAmount.Value)
		return
	}
	if b.UnconfirmedAmount.Value != 800000000000 {
		t.Errorf("test failed, expect unconfirmed balance %d, but got %d", 800000000000, b.UnspentAmount.Value)
		return
	}
}

func TestSpent(t *testing.T) {
	args := []string{"--modules=miner", "--modules=qitmeer", "--notls"}
	h, err := NewHarness(t, params.PrivNetParam.Params, args...)
	defer h.Teardown()

	if err != nil {
		t.Errorf("new harness failed: %v", err)
		return
	}
	err = h.Setup()
	if err != nil {
		t.Errorf("setup harness failed:%v", err)
		return
	}
	h.WaitWalletInit()

	if info, err := h.Client.NodeInfo(); err != nil {
		t.Errorf("test failed : %v", err)
		return
	} else {
		expect := "privnet"
		if info.Network != expect {
			t.Errorf("test failed, expect %v , but got %v", expect, info.Network)
			return
		}

	}
	GenerateBlock(t, h, 18)
	time.Sleep(10 * time.Second)
	b, err := h.wallet.Balance(types.MEERA)
	if err != nil {
		t.Errorf("test failed : %v", err)
		return
	}
	_, err = h.wallet.SendToAddress("RmV7i7JoomcHuQCVMN66SiTYUCkRtzQ6fSf", types.MEERA, 498)
	if err != nil {
		t.Errorf("test failed, %v", err)
	}
	GenerateBlock(t, h, 1)
	time.Sleep(10 * time.Second)
	b, err = h.wallet.Balance(types.MEERA)
	if err != nil {
		t.Errorf("test failed : %v", err)
		return
	}
	if b.SpendAmount.Value != 50000000000 {
		t.Errorf("test failed, expect spent balance %d, but got %d", 50000000000, b.SpendAmount.Value)
		return
	}
	GenerateBlock(t, h, 1)
	b, err = h.wallet.BalanceByAddr(types.MEERA, "RmV7i7JoomcHuQCVMN66SiTYUCkRtzQ6fSf")
	if err != nil {
		t.Errorf("test failed : %v", err)
		return
	}

	fmt.Println(b.LockAmount, b.UnspentAmount, b.TotalAmount, b.SpendAmount, b.UnconfirmedAmount)

}
