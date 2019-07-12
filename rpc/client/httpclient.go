package client

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/samuel/go-socks/socks"
)

// newHTTPClient returns a new HTTP client that is configured according to the
// proxy and TLS settings in the associated connection configuration.
func NewRPCClient(cfg *Config) (*RPCClient, error) {
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
	client := http.Client{
		Transport: &http.Transport{
			Dial:            dial,
			TLSClientConfig: tlsConfig,
		},
	}

	// Generate a request to the configured RPC server.
	protocol := "http"
	if !cfg.NoTLS {
		protocol = "https"
	}
	url := protocol + "://" + cfg.RPCServer

	rpcClient := &RPCClient{
		Cfg:    cfg,
		Client: &client,
		URL:    url,
	}

	return rpcClient, nil
}

type RPCClient struct {
	Cfg *Config

	Client *http.Client

	URL string
}

// sendPostRequest sends the marshalled JSON-RPC command using HTTP-POST mode
// to the server described in the passed config struct.  It also attempts to
// unmarshal the response as a JSON-RPC response and returns either the result
// field or the error field depending on whether or not there is an error.
func (rpcClient *RPCClient) sendPostRequest(path string, marshalledJSON []byte) ([]byte, error) {

	bodyReader := bytes.NewReader(marshalledJSON)

	httpRequest, err := http.NewRequest("POST", rpcClient.URL+"/"+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("sendPostRequest: htt.NewRequest err: %s", err)
	}
	httpRequest.Close = true
	httpRequest.Header.Set("Content-Type", "application/json")

	// Configure basic access authorization.
	httpRequest.SetBasicAuth(rpcClient.Cfg.RPCUser, rpcClient.Cfg.RPCPassword)

	httpResponse, err := rpcClient.Client.Do(httpRequest)
	if err != nil {
		return nil, fmt.Errorf("sendPostRequest: httpClient.Do err: %s", err)
	}

	// Read the raw bytes and close the response.
	respBytes, err := ioutil.ReadAll(httpResponse.Body)
	httpResponse.Body.Close()
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
	// 	fmt.Println(string(respBytes))
	// }

	// Unmarshal the response.
	var resp Response
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, fmt.Errorf("sendPostRequest: json.Unmarshal resData: %s", respBytes)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("sendPostRequest: resp.Error: %s,sendData: %s", respBytes, string(marshalledJSON))
	}
	return resp.Result, nil
}

type Response struct {
	Jsonrpc string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   *RPCError       `json:"error"`
	ID      *interface{}    `json:"id"`
}

// RPCErrorCode A specific type is used to help ensure the wrong errors aren't used.
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
func (rpcClient *RPCClient) makeRequestData(rpcVersion string, id interface{}, method string, params []interface{}) ([]byte, error) {
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

const rpcVersion = "1.0"

func (rcpClient *RPCClient) GetResString(method string, args []interface{}) (rs string, err error) {
	reqData, err := rcpClient.makeRequestData(rpcVersion, 1, method, args)
	if err != nil {
		err = fmt.Errorf("getResString [%s]: %s", method, err)
		return
	}

	resResult, err := rcpClient.sendPostRequest("", reqData)
	if err != nil {
		err = fmt.Errorf("getResString [%s]: %s", method, err)
		return
	}

	rs = string(resResult)
	return
}
