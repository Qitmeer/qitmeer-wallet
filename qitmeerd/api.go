package qitmeerd

import (
	"fmt"

	"github.com/Qitmeer/qitmeer/log"

	"github.com/Qitmeer/qitmeer-wallet/config"
	"github.com/Qitmeer/qitmeer-wallet/rpc/client"
	"github.com/Qitmeer/qitmeer-wallet/wallet"
)

// API to mgr qitmeerd
type API struct {
	cfg      *config.Config
	qitmeerd *Qitmeerd
}

// NewAPI api make
func NewAPI(cfg *config.Config, qitmeerd *Qitmeerd) *API {
	return &API{
		cfg:      cfg,
		qitmeerd: qitmeerd,
	}
}

// List list all qitmeerd
func (api *API) List() ([]*client.Config, error) {
	return api.cfg.Qitmeerds, nil
}

// Add a new qitmeerd conf
func (api *API) Add(name string,
	RPCServer string, RPCUser string, RPCPassword string,
	RPCCert string, NoTLS bool, TLSSkipVerify bool,
	Proxy string, ProxyUser string, ProxyPass string) error {

	for _, item := range api.cfg.Qitmeerds {
		if item.Name == name {
			return nil
		}
	}

	api.cfg.Qitmeerds = append(api.cfg.Qitmeerds, &client.Config{
		Name:          name,
		RPCUser:       RPCUser,
		RPCPassword:   RPCPassword,
		RPCServer:     RPCServer,
		RPCCert:       RPCCert,
		NoTLS:         NoTLS,
		TLSSkipVerify: TLSSkipVerify,
		Proxy:         Proxy,
		ProxyUser:     ProxyUser,
		ProxyPass:     ProxyPass,
	})

	api.cfg.Save(api.cfg.ConfigFile)

	return nil
}

// Del a qitmeerd conf
func (api *API) Del(name string) error {
	nameP := -1
	for i, item := range api.cfg.Qitmeerds {
		if item.Name == name {
			nameP = i
			break
		}
	}
	if nameP == -1 {
		return nil
	}

	api.cfg.Qitmeerds = append(api.cfg.Qitmeerds[:nameP], api.cfg.Qitmeerds[nameP+1:]...)
	api.cfg.Save(api.cfg.ConfigFile)
	return nil
}

// Update qitmeerd conf
func (api *API) Update(name string,
	RPCServer string, RPCUser string, RPCPassword string,
	RPCCert string, NoTLS bool, TLSSkipVerify bool,
	Proxy string, ProxyUser string, ProxyPass string) error {

	var updateQitmeerd *client.Config
	for _, item := range api.cfg.Qitmeerds {
		if item.Name == name {
			updateQitmeerd = item
			break
		}
	}

	updateQitmeerd.RPCUser = RPCUser
	updateQitmeerd.RPCPassword = RPCPassword
	updateQitmeerd.RPCServer = RPCServer
	updateQitmeerd.RPCCert = RPCCert
	updateQitmeerd.NoTLS = NoTLS
	updateQitmeerd.TLSSkipVerify = TLSSkipVerify
	updateQitmeerd.Proxy = Proxy
	updateQitmeerd.ProxyUser = ProxyUser
	updateQitmeerd.ProxyPass = ProxyPass

	api.cfg.Save(api.cfg.ConfigFile)

	return nil
}

// Reset qitmeerd rpc client
func (api *API) Reset(name string) error {

	if api.cfg.QitmeerdSelect == name {
		log.Trace("not reset qitmeerd,it eq")
		return nil
	}

	var resetQitmeerd *client.Config
	for _, item := range api.cfg.Qitmeerds {
		if item.Name == name {
			resetQitmeerd = item
			break
		}
	}

	if resetQitmeerd == nil {
		return fmt.Errorf("qitmeerd %s not found", name)
	}

	htpc, err := wallet.NewHtpcByCfg(resetQitmeerd)
	if err != nil {
		return fmt.Errorf("make rpc clent error: %s", err.Error())
	}

	api.cfg.QitmeerdSelect = name
	api.qitmeerd.Status.CurrentName = name
	// update wallet httpclient
	api.qitmeerd.Wt.Httpclient = htpc

	return nil
}

// Status get qitmeerd stats
func (api *API) Status() (*Status, error) {
	return api.qitmeerd.Status, nil
}

//Status qitmeerd status
type Status struct {
	CurrentName  string //current qitmeerd name
	err          string //
	MainOrder    uint32
	MainHeight   uint32
	Blake2bdDiff string // float64
	CuckarooDiff float64
	CuckatooDiff float64
}
