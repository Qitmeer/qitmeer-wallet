module github.com/Qitmeer/qitmeer-wallet

go 1.15

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/Qitmeer/qitmeer v0.10.0-dev.0.20210406085400-b7642766e7b2
	github.com/deckarep/golang-set v1.7.1
	github.com/jessevdk/go-flags v1.5.0
	github.com/jrick/logrotate v1.0.0
	github.com/julienschmidt/httprouter v1.3.0
	github.com/mattn/go-colorable v0.1.8
	github.com/peterh/liner v1.2.1
	github.com/rakyll/statik v0.1.7
	github.com/samuel/go-socks v0.0.0-20130725190102-f6c5f6a06ef6
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.7.1
	go.etcd.io/bbolt v1.3.5
	golang.org/x/crypto v0.0.0-20210421170649-83a5a9bb288b
	golang.org/x/net v0.0.0-20210421230115-4e50805a0758
)

replace (
	github.com/Qitmeer/qitmeer v0.10.0-dev.0.20210406085400-b7642766e7b2 => ../qitmeer-jamesvan2019
)