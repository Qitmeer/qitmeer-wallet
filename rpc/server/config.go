package server

// Config Server config
type Config struct {
	RPCUser       string `short:"u" long:"rpcuser" description:"Username for RPC connections"`
	RPCPass       string `short:"P" long:"rpcpass" default-mask:"-" description:"Password for RPC connections"`
	RPCCert       string `long:"rpccert" description:"File containing the certificate file"`
	RPCKey        string `long:"rpckey" description:"File containing the certificate key"`
	RPCMaxClients int    `long:"rpcmaxclients" description:"Max number of RPC clients for standard connections"`
	DisableRPC    bool   `long:"norpc" description:"Disable built-in RPC server -- NOTE: The RPC server is disabled by default if no rpcuser/rpcpass or rpclimituser/rpclimitpass is specified"`
	DisableTLS    bool   `long:"notls" description:"Disable TLS for the RPC server -- NOTE: This is only allowed if the RPC server is bound to localhost"`
}
