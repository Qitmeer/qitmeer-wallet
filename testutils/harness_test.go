// Copyright (c) 2020 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package testutils

import (
	"github.com/Qitmeer/qitmeer/params"
	"sync"
	"testing"
	"time"
)

func TestHarness(t *testing.T) {
	h, err := NewHarness(t, params.PrivNetParam.Params)
	if err != nil {
		t.Errorf("create new test harness instance failed %v", err)
	}
	if err := h.Setup(); err != nil {
		t.Errorf("setup test harness instance failed %v", err)
	}

	h2, err := NewHarness(t, params.PrivNetParam.Params)
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
			NewHarness(t, params.PrivNetParam.Params)
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
	time.Sleep(500 * time.Millisecond)

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
	b, err := h.wallet.Balance("MEER")
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
	time.Sleep(500 * time.Millisecond)

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
	GenerateBlock(t, h, 20)
	time.Sleep(10 * time.Second)

	b, err := h.wallet.Balance("MEER")
	if err != nil {
		t.Errorf("test failed : %v", err)
		return
	}
	if b.UnspentAmount.Value != 200000000000 {
		t.Errorf("test failed, expect unspent balance %d, but got %d", 200000000000, b.UnspentAmount.Value)
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
	time.Sleep(500 * time.Millisecond)

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
	GenerateBlock(t, h, 20)
	time.Sleep(10 * time.Second)

	b, err := h.wallet.Balance("MEER")
	if err != nil {
		t.Errorf("test failed : %v", err)
		return
	}
	_, err = h.wallet.SendToAddress("RmV7i7JoomcHuQCVMN66SiTYUCkRtzQ6fSf", "MEER", 2000)
	if err != nil {
		t.Errorf("test failed, %v", err)
	}
	b, err = h.wallet.Balance("MEER")
	if err != nil {
		t.Errorf("test failed : %v", err)
		return
	}
	if b.UnspentAmount.Value != 0 {
		t.Errorf("test failed, expect unspent balance %d, but got %d", 0, b.UnspentAmount.Value)
		return
	}
	if b.SpendAmount.Value != 200000000000 {
		t.Errorf("test failed, expect spent balance %d, but got %d", 200000000000, b.UnspentAmount.Value)
		return
	}
	if b.TotalAmount.Value != 800000000000 {
		t.Errorf("test failed, expect total balance %d, but got %d", 800000000000, b.UnspentAmount.Value)
		return
	}
}
