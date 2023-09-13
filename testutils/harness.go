// Copyright (c) 2020 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package testutils

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/Qitmeer/qng/core/protocol"
	"github.com/Qitmeer/qng/params"
	"github.com/ethereum/go-ethereum/ethclient"
)

const DefaultMaxRpcConnRetries = 10

var (
	// harness main-process id which shared for all harness instances
	harnessMainProcessId = os.Getpid()

	// the private harness instances map contains all initialized harnesses
	// which returned by the NewHarness func. and the instance will delete itself
	// from the map after the Teardown func has been called
	harnessInstances = make(map[string]*Harness)
	// protect the global harness state variables
	harnessStateMutex sync.RWMutex
)

// Harness manage an embedded qitmeer node process for running the rpc driven
// integration tests. The active qitmeer node will typically be run in privnet
// mode in order to allow for easy block generation. Harness handles the node
// start/shutdown and any temporary directories need to be created.
type Harness struct {
	// Harness id
	id int
	// the temporary directory created when the Harness instance initialized
	// its also used as the unique id of the harness instance, its in the
	// format of `test-harness-<num>-*`
	instanceDir string
	// the qitmeer node process
	Node *node
	// the rpc client to the qitmeer node in the Harness instance.
	Client    *Client
	evmClient *ethclient.Client
	// the maximized attempts try to establish the rpc connection
	maxRpcConnRetries int
	// Notifier use rpc/client with web-socket notification support
	wallet *Wallet
}

func (h *Harness) Id() string {
	return strconv.Itoa(h.id) + "_" + h.instanceDir
}

// Setup func initialize the test state.
// 1. start the qitmeer node according to the net params
// 2. setup the rpc clint so that the rpc call can be sent to the node
// 3. generate a test block dag by configuration (optional, may empty dag by config)
func (h *Harness) Setup() error {
	// start up the qitmeer node
	if err := h.Node.start(); err != nil {
		return err
	}
	// setup the rpc client
	if err := h.connectRPCClient(); err != nil {
		return err
	}
	if err := h.wallet.Start(); err != nil {
		return err
	}
	return nil
}

// connectRPCClient attempts to establish an RPC connection to the Harness instance.
// If the initial attempt fails, this function will retry h.maxRpcConnRetries times,
// this function returns with an error if all retries failed.
func (h *Harness) connectRPCClient() error {
	var client *Client
	var err error
	var http = "https://"
	url, user, pass, notls := h.Node.config.rpclisten, h.Node.config.rpcuser, h.Node.config.rpcpass, h.Node.config.rpcnotls
	certs := h.Node.config.certFile
	if notls {
		http = "http://"
	}
	for i := 0; i < h.maxRpcConnRetries; i++ {
		if client, err = Dial(http+url, user, pass, certs); err != nil {
			time.Sleep(time.Duration(i) * time.Second)
			continue
		}
		break
	}
	if client == nil || err != nil {
		return fmt.Errorf("failed to establish rpc client connection: %v", err)
	}
	for i := 0; i < h.maxRpcConnRetries; i++ {
		if h.evmClient, err = ethclient.Dial("http://127.0.0.1:" + h.Node.config.evmlisten); err != nil {
			time.Sleep(time.Duration(i) * time.Second)
			continue
		}
		break
	}
	if h.evmClient == nil || err != nil {
		return fmt.Errorf("failed to establish evm client connection: %v", err)
	}
	h.Client = client
	return nil
}

// Teardown func the concurrent safe wrapper of teardown func
func (h *Harness) Teardown() error {
	harnessStateMutex.Lock()
	defer harnessStateMutex.Unlock()
	return h.teardown()
}

func (h *Harness) WaitWalletInit() {
	for {
		if h.wallet.wallet.UploadRun {
			break
		} else {
			time.Sleep(500 * time.Millisecond)
		}
	}
}

// teardown func stop the running test, stop the rpc client shutdown the node,
// kill any related processes if need and clean up the temporary data folder
// NOTE: the func is NOT concurrent safe. see also the Teardown func
func (h *Harness) teardown() error {
	if err := h.Node.stop(); err != nil {
		return err
	}
	h.wallet.Stop()
	if err := os.RemoveAll(h.instanceDir); err != nil {
		return err
	}

	delete(harnessInstances, h.instanceDir)
	return nil
}

// NewHarness func creates an new instance of test harness with provided network params.
// The args is the arguments list that are used when setup a qitmeer node. In the most
// case, it should be set to nil if no extra args need to add on the default starting up.
func NewHarness(t *testing.T, params *params.Params, args ...string) (*Harness, error) {
	return NewHarnessWithMnemonic(t, "", "", false, params, args...)
}

// NewHarness func creates an new instance of test harness with provided network params.
// The args is the arguments list that are used when setup a qitmeer node. In the most
// case, it should be set to nil if no extra args need to add on the default starting up.
func NewHarnessWithMnemonic(t *testing.T, mnemonic, path string, usePkAddr bool, params *params.Params, args ...string) (*Harness, error) {
	harnessStateMutex.Lock()
	defer harnessStateMutex.Unlock()
	id := len(harnessInstances)
	// create temporary folder
	testDir, err := ioutil.TempDir("", "test-harness-"+strconv.Itoa(int(time.Now().UnixNano())+id)+"-*")
	if err != nil {
		return nil, err
	}

	// setup network type
	extraArgs := []string{}
	switch params.Net {
	case protocol.MainNet:
		//do nothing for mainnet which is by default
	case protocol.MixNet:
		extraArgs = append(extraArgs, "--mixnet")
	case protocol.TestNet:
		extraArgs = append(extraArgs, "--testnet")
	case protocol.PrivNet:
		extraArgs = append(extraArgs, "--privnet")
	default:
		return nil, fmt.Errorf("unknown network type %v", params.Net)
	}

	// force using notls since web-socket not support tls yet.
	// extraArgs = append(extraArgs, "--notls")

	// create node config & initialize the node process

	walletCfg := newWalletConfig(testDir)
	// use auto-genereated p2p/rpc port settings instead of default
	p2pListen, rpcListen, evmListen, evmWS := genListenArgs()

	walletCfg.QServer = rpcListen
	wallet, err := NewWallet(walletCfg, params.Net, mnemonic, path)
	if err != nil {
		return nil, err
	}
	address, err := wallet.GenerateAddress(usePkAddr)
	if err != nil {
		return nil, err
	}
	fmt.Printf("wallet address is %s\n", address)
	extraArgs = append(extraArgs, args...)
	extraArgs = append(extraArgs, fmt.Sprintf("--miningaddr=%s", address))

	config := newNodeConfig(testDir, extraArgs)
	config.listen, config.rpclisten = p2pListen, rpcListen
	config.evmlisten = evmListen
	config.evmWSlisten = evmWS
	// create node
	newNode, err := newNode(t, config)
	if err != nil {
		return nil, err
	}
	h := Harness{
		id:                id,
		instanceDir:       testDir,
		Node:              newNode,
		maxRpcConnRetries: DefaultMaxRpcConnRetries,
		wallet:            wallet,
	}
	harnessInstances[h.instanceDir] = &h
	return &h, nil
}

// TearDownAll func teardown all Harness Instances
func TearDownAll() error {
	harnessStateMutex.Lock()
	defer harnessStateMutex.Unlock()
	for _, h := range harnessInstances {
		if err := h.teardown(); err != nil {
			return err
		}
	}
	return nil
}

// AllHarnesses func get all Harness instances
func AllHarnesses() []*Harness {
	harnessStateMutex.RLock()
	defer harnessStateMutex.RUnlock()
	all := make([]*Harness, 0, len(harnessInstances))
	for _, h := range harnessInstances {
		all = append(all, h)
	}
	return all
}

const (
	// the minimum and maximum p2p and rpc port numbers used by a test harness.
	minP2PPort   = 18200               // 18200 The min is inclusive
	maxP2PPort   = minP2PPort + 1000   // 19199 The max is exclusive
	minRPCPort   = maxP2PPort          // 19200
	maxRPCPort   = minRPCPort + 1000   // 30199
	minEVMPort   = maxRPCPort          // 30200
	maxEVMPort   = minEVMPort + 1000   // 31199
	minEVMWSPort = maxEVMPort          // 31200
	maxEVMWSPort = minEVMWSPort + 1000 // 32199
)

var argsLock sync.Mutex

// GenListenArgs returns auto generated args for p2p listen and rpc listen in the format of
// ["--listen=127.0.0.1:12345", --rpclisten=127.0.0.1:12346"].
// in order to support multiple test node running at the same time.
func genListenArgs() (string, string, string, string) {
	argsLock.Lock()
	defer argsLock.Unlock()
	localhost := "127.0.0.1"
	genPort := func(min, max int) string {
		rand.Seed(time.Now().UnixNano() + int64(len(harnessInstances)))
		return fmt.Sprintf("%d", rand.Intn(max-min)+min)
		// port := min + len(harnessInstances) + (42 * harnessMainProcessId % (max - min))
		// return strconv.Itoa(port)
	}
	p2p := net.JoinHostPort(localhost, genPort(minP2PPort, maxP2PPort))
	rpc := net.JoinHostPort(localhost, genPort(minRPCPort, maxRPCPort))
	evm := genPort(minEVMPort, maxEVMPort)
	evmWS := genPort(minEVMWSPort, maxEVMWSPort)
	return p2p, rpc, evm, evmWS
}
