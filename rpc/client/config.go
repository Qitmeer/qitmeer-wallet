package client

// Config file
type Config struct {
	RPCUser       string
	RPCPassword   string
	RPCServer     string
	RPCCert       string
	NoTLS         bool
	TLSSkipVerify bool

	Proxy     string
	ProxyUser string
	ProxyPass string
}
