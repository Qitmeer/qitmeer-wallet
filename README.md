## qitmeer-wallet
The command-line wallet of the Qitmeer network


## download or build

you can download from [release](https://github.com/Qitmeer/qitmeer-wallet/releases)

if you have golang environment, you can build it by yourself

```
git clone https://github.com/Qitmeer/qitmeer-wallet.git

go build

```

## Usage

```
Usage:
  qitmeer-wallet [command]

Available Commands:
  help        Help about any command
  qc          qitmeer wallet command
  qx          qx util

Flags:
  -h, --help   help for qitmeer-wallet


Usage:
  qitmeer-wallet qc [command]

Available Commands:
  console               console
  create                create
  createnewaccount      create new account
  getaddressesbyaccount get addresses by account
  getbalance            getbalance
  getlisttxbyaddr       get all transactions for address
  importprivkey         import prikey
  listaccountsbalance   list Accounts Balance
  sendtoaddress         send transaction
  syncheight            Get the number of local synchronization blocks
  updateblock           Update local block data
  web                   web

Flags:
  -a, --appdatadir string       wallet db path
  -c, --configfile string       config file (default "config.toml")
      --create                  Create a new wallet
  -d, --debuglevel string       Logging level {trace, debug, info, warn, error, critical} (default "info")
  -h, --help                    help for qc
      --listeners stringArray   rpc listens (default [127.0.0.1:38130])
  -l, --logdir string           log data path
  -n, --network string          network (default "testnet")
  -P, --pubwalletpass string    data encryption password (default "public")
  -p, --qpass string            qitmeer node password (default "123456")
  -s, --qserver string          qitmeer node server (default "127.0.0.1:8030")
  -u, --quser string            qitmeer node user (default "admin")
      --rpcPass string          rpc pass,default by random (default "yKR6RvDiwgSEt9he9GJUzWnk")
      --rpcUser string          rpc user,default by random (default "l6Jm57B5")
      --ui                      Start Wallet with RPC and webUI interface (default true)

Use "qitmeer-wallet qc [command] --help" for more information about a command.


Usage:
  qitmeer-wallet qx [command]

Available Commands:
  generatemnemonic generate mnemonic
  mnemonictoaddr   mnemonic to address
  mnemonictoseed   mnemonic to seed
  pritoaddr        private key to address
  pritopub         private key to public key
  pubtoaddr        public key to address
  seedtoaddr       seed to address
  seedtopri        Seed private key

Flags:
  -h, --help   help for qx

Use "qitmeer-wallet qx [command] --help" for more information about a command.

```

## How to use qitmeer-wallet console command model

1:  configure the config.toml file in the root directory

    ```
        #configFile="" #Your config.toml profile directory
        #appDataDir="" # Your DB storage path
        #logDir="" # log path
        #network="mainnet" #network mainnet,testnet,privnet default testnet
        network="testnet"
        #network="privnet"
        #Qitmeerd
        QServer=""
        QUser=""
        QPass=""
        WalletPass="public" #Wallet encryption code default public
    
    ```

2:  create a  wallet

    ```
    ./qitmeer-wallet qc create 
    
    ```
    
    `#output`
    
    ```
     # Wallet Password
     Enter the private passphrase for your new wallet:
     Confirm passphrase:
     # Wallet data encryption password, default public
     Do you want to add an additional layer of encryption for public data? (n/no/y/yes) [no]:y
     Enter the public passphrase for your new wallet:
     Confirm passphrase:
     NOTE: Use the -- wallet pass option to configure your public passphrase.
     PubPass: 123
     # Whether to import wallet seeds
     Do you have an existing wallet seed you want to use? (n/no/y/yes) [no]: n
     Your wallet generation seed is:
     17e28af99e36ff4527c95f91e13d3ecd82349864d23b9ff2d4f9d446ea078291
     IMPORTANT: Keep the seed in a safe place as you
     Will NOT be able to restore your wallet without it.
     Please keep in mind that any who has access
     To the seed can also restore your wallet thus
     Give them access to all your funds, so it is
     Imperative that you keep it in a secure location.
     Once you have stored the seed in a safe and secure location, enter "OK" to continue: OK
     Creating the wallet...
     INFO [0021] Opened Wallet
     Pri: 6eb6bbcd7ded317abc4ed5e373c2c8630dc4ad069470ad7ae72f5fb854423006
     INFO [0022] Imported payment address TmmC1hUN5A2RJzX9ZWFZqHaDbKUf6NaA4D
     The wallet has been created successfully.
     ```

3:  update db (update to the specified block or update all,syncheight can view the current DB synchronization height)
    
    ./qitmeer-wallet qc updateblock
    ./qitmeer-wallet qc updateblock 130 
    
    
    
4:  when creating a wallet, you can import seeds or import private keys using importprivkey.

    `./qitmeer-wallet qc importprivkey 6eb6bbcd7ded317abc4ed5e373c2c8630dc4ad069470ad7ae72f5fb854423006` 
    
    `#output`
    
    `ImportPrivKey: OK`

    
5:  must use updateback to see the balance change after the transfer transaction

    `./qitmeer-wallet qc sendtoaddress TmbsdsjwzuGboFQ9GcKg6EUmrr3tokzozyF 0.01 youpassword`
    
    `#output`
    
    `0e441ecee44defe28711103eef0cc3d01c187c257738150869c032fbbf96d4c9`

    `./qitmeer-wallet qc updateback`
