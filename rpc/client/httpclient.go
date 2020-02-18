package client

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/Qitmeer/qitmeer/log"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/samuel/go-socks/socks"
)

// newHTTPClient returns a new HTTP client that is configured according to the
// proxy and TLS settings in the associated connection configuration.
func newHTTPClient(cfg *Config) (*http.Client, error) {
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

	timeout, _ := time.ParseDuration("30s")

	// Create and return the new HTTP client potentially configured with a
	// proxy and TLS.
	client := http.Client{
		Transport: &http.Transport{
			Dial:            dial,
			TLSClientConfig: tlsConfig,
		},
		Timeout: timeout,
	}
	return &client, nil
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
func MakeRequestData(rpcVersion string, id interface{}, method string, params []interface{}) ([]byte, error) {
	// default to JSON-RPC 1.0 if RPC type is not specified
	if rpcVersion != "2.0" && rpcVersion != "1.0" {
		rpcVersion = "1.0"
	}
	if !IsValidIDType(id) {
		return nil, fmt.Errorf("requestData err: %T is invalid", id)
	}

	rawParams := make([]json.RawMessage, 0, len(params))
	for _, param := range params {
		marshalledParam, err := json.Marshal(param)
		if err != nil {
			return nil, fmt.Errorf("requestData err: Marshal param err: %s ", err)
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
		return nil, fmt.Errorf("requestData err: Marshal err: %s", err)
	}

	log.Trace("Post data: ", string(reqData))

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
