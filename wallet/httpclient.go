package wallet

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/Qitmeer/qitmeer/log"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"

	qJson "github.com/Qitmeer/qitmeer/core/json"

	"github.com/Qitmeer/qitmeer-wallet/config"
	"github.com/Qitmeer/qitmeer-wallet/rpc/client"

	"github.com/samuel/go-socks/socks"
)

type httpConfig struct {
	RPCUser       string
	RPCPassword   string
	RPCServer     string
	RPCCert       string
	NoTLS         bool
	TLSSkipVerify bool

	Proxy      string
	ProxyUser  string
	ProxyPass  string
	httpClient *http.Client
}

const (
	strIntBase      = 10
	strIntBitSize32 = 32
	syncDiffNum     = 200
)

// NewHtpc make qitmeerd http client
func NewHtpc() (*httpConfig, error) {

	if config.Cfg.QitmeerdSelect != "" {
		var qitmeerd *client.Config
		for _, item := range config.Cfg.Qitmeerds {
			if item.Name == config.Cfg.QitmeerdSelect {
				qitmeerd = item
				break
			}
		}
		if qitmeerd != nil {
			return NewHtpcByCfg(qitmeerd)
		}
	}

	h := &httpConfig{
		RPCUser:       config.Cfg.QUser,
		RPCPassword:   config.Cfg.QPass,
		RPCServer:     config.Cfg.QServer,
		RPCCert:       config.Cfg.QCert,
		NoTLS:         config.Cfg.QNoTLS,
		TLSSkipVerify: config.Cfg.QTLSSkipVerify,
		Proxy:         config.Cfg.QProxy,
		ProxyUser:     config.Cfg.QProxyUser,
		ProxyPass:     config.Cfg.QProxyPass,
	}
	c, err := newHTTPClient(h)
	if err != nil {
		return nil, err
	}
	h.httpClient = c
	return h, nil
}

// NewHtpcByCfg new httpConfig by cfg
func NewHtpcByCfg(cfg *client.Config) (*httpConfig, error) {
	h := &httpConfig{
		RPCUser:       cfg.RPCUser,
		RPCPassword:   cfg.RPCPassword,
		RPCServer:     cfg.RPCServer,
		RPCCert:       cfg.RPCCert,
		NoTLS:         cfg.NoTLS,
		TLSSkipVerify: cfg.TLSSkipVerify,
		Proxy:         cfg.Proxy,
		ProxyUser:     cfg.ProxyUser,
		ProxyPass:     cfg.ProxyPass,
	}
	c, err := newHTTPClient(h)
	if err != nil {
		return nil, err
	}
	h.httpClient = c
	return h, nil
}

// newHTTPClient returns a new HTTP client that is configured according to the
// proxy and TLS settings in the associated connection configuration.
func newHTTPClient(cfg *httpConfig) (*http.Client, error) {
	// Configure proxy if needed.
	var dial func(network, addr string) (net.Conn, error)
	if cfg.Proxy != "" {
		proxy := &socks.Proxy{
			Addr:     cfg.Proxy,
			Username: cfg.ProxyUser,
			Password: cfg.ProxyPass,
		}
		dial = func(network, addr string) (net.Conn, error) {
			c, err := proxy.Dial(network, addr)
			if err != nil {
				return nil, err
			}
			return c, nil
		}
	}

	// Configure TLS if needed.
	var tlsConfig *tls.Config
	if !cfg.NoTLS {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: cfg.TLSSkipVerify,
		}
		if !cfg.TLSSkipVerify && cfg.RPCCert != "" {
			pem, err := ioutil.ReadFile(cfg.RPCCert)
			if err != nil {
				return nil, err
			}

			pool := x509.NewCertPool()
			if ok := pool.AppendCertsFromPEM(pem); !ok {
				return nil, fmt.Errorf("invalid certificate file: %v",
					cfg.RPCCert)
			}
			tlsConfig.RootCAs = pool
		}
	}

	// Create and return the new HTTP client potentially configured with a
	// proxy and TLS.
	c := http.Client{
		Transport: &http.Transport{
			Dial:            dial,
			TLSClientConfig: tlsConfig,
		},
	}
	return &c, nil
}

func (cfg *httpConfig) CheckSyncUpdate(localheight int64) (bool, error) {
	var params []interface{}
	str, err := cfg.getResString("getBlockCount", params)
	if err != nil {
		return false, err
	}
	blockHeight, err := strconv.ParseInt(str, strIntBase, strIntBitSize32)
	if err != nil {
		return false, err
	}
	log.Trace(fmt.Sprintf("blockheight:%v,localheight:%v", blockHeight, localheight))
	if uint32(blockHeight-localheight) < (config.Cfg.Confirmations + syncDiffNum) {
		return true, nil
	} else {
		return false, fmt.Errorf("db Update incomplete")
	}
}

func (cfg *httpConfig) getblockCount() (string, error) {
	var params []interface{}
	return cfg.getResString("getBlockCount", params)
}
func (cfg *httpConfig) getMempool() (string, error) {
	params := []interface{}{"", false}
	return cfg.getResString("getMempool", params)
}
func (cfg *httpConfig) getRawTransaction(txhash string) (string, error) {
	params := []interface{}{txhash, true}
	return cfg.getResString("getRawTransaction", params)
}
func (cfg *httpConfig) getBlockhash(i int64) (string, error) {
	params := []interface{}{i}
	str, err := cfg.getResString("getBlockhash", params)
	return strings.Replace(str, "\"", "", -1), err
}
func (cfg *httpConfig) getBlock(hash string, isDetail bool) (string, error) {
	params := []interface{}{hash, isDetail}
	return cfg.getResString("getBlock", params)
}
func (cfg *httpConfig) getBlockByOrder(i int64) ([]byte, error) {
	params := []interface{}{i, true}
	return cfg.getResByte("getBlockByOrder", params)
}
func (cfg *httpConfig) isBlue(blockHash string) (bool, error) {
	params := []interface{}{blockHash}
	isBlue, err := cfg.getResString("isBlue", params)
	if err == nil {
		if isBlue != "1" {
			return false, nil
		} else {
			return true, nil
		}
	}
	return false, err
}
func (cfg *httpConfig) SendRawTransaction(tx string, allowHighFees bool) (string, error) {
	params := []interface{}{tx, allowHighFees}
	return cfg.getResString("sendRawTransaction", params)
}

func (cfg *httpConfig) GetNodeInfo() (*qJson.InfoNodeResult, error) {
	var params []interface{}
	buf, err := cfg.getResByte("getNodeInfo", params)
	if err != nil {
		return nil, err
	}
	nodeInfo := &qJson.InfoNodeResult{}
	err = json.Unmarshal(buf, nodeInfo)
	if err != nil {
		return nil, err
	}

	return nodeInfo, nil
}

// sendPostRequest sends the marshalled JSON-RPC command using HTTP-POST mode
// to the server described in the passed config struct.  It also attempts to
// unmarshal the response as a JSON-RPC response and returns either the result
// field or the error field depending on whether or not there is an error.
func (cfg *httpConfig) sendPostRequest(marshalledJSON []byte) ([]byte, error) {
	// Generate a request to the configured RPC server.
	protocol := "http"
	if !cfg.NoTLS {
		protocol = "https"
	}
	url := protocol + "://" + cfg.RPCServer
	// if cfg.PrintJSON {
	// 	log.Info(string(marshalledJSON))
	// }
	bodyReader := bytes.NewReader(marshalledJSON)
	httpRequest, err := http.NewRequest("POST", url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("sendPostRequest: htt.NewRequest err: %s", err)
	}
	httpRequest.Close = true
	httpRequest.Header.Set("Content-Type", "application/json")

	// Configure basic access authorization.
	httpRequest.SetBasicAuth(cfg.RPCUser, cfg.RPCPassword)

	httpResponse, err := cfg.httpClient.Do(httpRequest)
	if err != nil {
		return nil, fmt.Errorf("sendPostRequest: httpClient.Do err: %s", err)
	}

	// Read the raw bytes and close the response.
	respBytes, err := ioutil.ReadAll(httpResponse.Body)
	_ = httpResponse.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("sendPostRequest: reading json reply: err: %v", err)
	}

	// Handle unsuccessful HTTP responses
	if httpResponse.StatusCode < 200 || httpResponse.StatusCode >= 300 {
		// Generate a standard error to return if the server body is
		// empty.  This should not happen very often, but it's better
		// than showing nothing in case the target server has a poor
		// implementation.
		if len(respBytes) == 0 {
			return nil, fmt.Errorf("%d %s", httpResponse.StatusCode,
				http.StatusText(httpResponse.StatusCode))
		}
		return nil, fmt.Errorf("sendPostRequest: StatusCode: %s", respBytes)
	}

	// If requested, print raw json response.
	// if cfg.PrintJSON {
	// 	log.Info(string(respBytes))
	// }

	// Unmarshal the response.
	var resp Response
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, fmt.Errorf("sendPostRequest: json.Unmarshal resData: %s", respBytes)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("sendPostRequest: resp.Error: %s", respBytes)
		//return nil, fmt.Errorf("sendPostRequest: resp.Error: %s,sendData: %s", respBytes, string(marshalledJSON))
	}
	return resp.Result, nil
}

type Response struct {
	Jsonrpc string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   *RPCError       `json:"error"`
	ID      *interface{}    `json:"id"`
}

// A specific type is used to help ensure the wrong errors aren't used.
type RPCErrorCode int

// RPCError represents an error that is used as a part of a JSON-RPC Response
// object.
type RPCError struct {
	Code    RPCErrorCode `json:"code,omitempty"`
	Message string       `json:"message,omitempty"`
}

func (e RPCError) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

//Request json req
type Request struct {
	Jsonrpc string            `json:"jsonrpc"`
	Method  string            `json:"method"`
	Params  []json.RawMessage `json:"params"`
	ID      interface{}       `json:"id"`
}

//makeRequestData
func makeRequestData(rpcVersion string, id interface{}, method string, params []interface{}) ([]byte, error) {
	// default to JSON-RPC 1.0 if RPC type is not specified
	if rpcVersion != "2.0" && rpcVersion != "1.0" {
		rpcVersion = "1.0"
	}
	if !IsValidIDType(id) {
		return nil, fmt.Errorf("makeRequestData: %T is invalid", id)
	}

	rawParams := make([]json.RawMessage, 0, len(params))
	for _, param := range params {
		marshalledParam, err := json.Marshal(param)
		if err != nil {
			return nil, fmt.Errorf("makeRequestData: Marshal param err: %s ", err)
		}
		rawMessage := json.RawMessage(marshalledParam)
		rawParams = append(rawParams, rawMessage)
	}

	req := Request{
		Jsonrpc: rpcVersion,
		ID:      id,
		Method:  method,
		Params:  rawParams,
	}

	reqData, err := json.Marshal(&req)
	if err != nil {
		return nil, fmt.Errorf("makeRequestData: Marshal err: %s", err)
	}
	return reqData, nil
}

//IsValidIDType id string number
func IsValidIDType(id interface{}) bool {
	switch id.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64,
		string,
		nil:
		return true
	default:
		return false
	}
}

var rpcVersion string = "1.0"

func (cfg *httpConfig) getResString(method string, args []interface{}) (rs string, err error) {
	reqData, err := makeRequestData(rpcVersion, 1, method, args)
	if err != nil {
		err = fmt.Errorf("getResString [%s]: %s", method, err)
		return
	}

	resResult, err := cfg.sendPostRequest(reqData)
	if err != nil {
		//err = fmt.Errorf("getResString [%s]: %s", method, err)
		return
	}

	rs = string(resResult)
	//log.Info("rs:",rs)
	return rs, err
}
func (cfg *httpConfig) getResByte(method string, args []interface{}) (rs []byte, err error) {
	reqData, err := makeRequestData(rpcVersion, 1, method, args)
	if err != nil {
		err = fmt.Errorf("getResString [%s]: %s", method, err)
		return
	}

	resResult, err := cfg.sendPostRequest(reqData)
	if err != nil {
		return nil, err
	}
	rs = resResult
	return rs, err
}
