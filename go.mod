module github.com/Qitmeer/qitmeer-wallet

go 1.12

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/Qitmeer/qitmeer v0.10.0-dev.0.20210323064630-6fff2ca01478
	github.com/davecgh/go-spew v1.1.1
	github.com/deckarep/golang-set v1.7.1
	github.com/jrick/logrotate v1.0.0
	github.com/julienschmidt/httprouter v1.2.0
	github.com/mattn/go-colorable v0.1.7
	github.com/peterh/liner v1.1.0
	github.com/rakyll/statik v0.1.7
	github.com/samuel/go-socks v0.0.0-20130725190102-f6c5f6a06ef6
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.3.2
	go.etcd.io/bbolt v1.3.3
	golang.org/x/crypto v0.0.0-20200510223506-06a226fb4e37
	golang.org/x/net v0.0.0-20200222125558-5a598a2470a0
)

replace (
	go.etcd.io/bbolt => github.com/etcd-io/bbolt v1.3.3
	golang.org/x/crypto => github.com/golang/crypto v0.0.0-20190701094942-4def268fd1a4
	golang.org/x/net => github.com/golang/net v0.0.0-20190628185345-da137c7871d7
	golang.org/x/sync => github.com/golang/sync v0.0.0-20190423024810-112230192c58
	golang.org/x/sys => github.com/golang/sys v0.0.0-20190710143415-6ec70d6a5542
	golang.org/x/text => github.com/golang/text v0.3.2
	golang.org/x/tools => github.com/golang/tools v0.0.0-20190710184609-286818132824
	gopkg.in/yaml.v2 => github.com/go-yaml/yaml v2.1.0+incompatible
)
