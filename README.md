# qitmeer-wallet
The command-line wallet of the Qitmeer network


# download or build

you can download from [release](https://github.com/HalalChain/qitmeer-wallet/releases)

if you have golang environment, you can build it by yourself

```

git clone https://github.com/HalalChain/qitmeer-wallet.git

cd qitmeer-wallet/cmds/qitwallet

go build

```

# usage

There are two ways to use it：RPC interface and Web interface.


## RPC interface

```
# start qitmeer-wallet

./qitwallet -n testnet --listens=127.0.0.1:38130 --rpcUser admin --rpcPass 123

# default RPC port is 38130

# example

#curl

curl -i -X POST -H 'Content-type':'application/json' --user uid:pwd -d '{"jsonrpc": "2.0","method": "getBalance","params": ["your-address"],"id": 1}' http://127.0.0.1:38130

```

you can also use [qitmeer-cli](https://github.com/HalalChain/qitmeer-cli) to access the qitmeer-wallet RPC interface.

```
./qitmeer-cli getbalance your-address
```

## Web interface

```
# start qitmeer-wallet

./qitmeer-wallet --web

# this will open http://127.0.0.1:38130 in your web browser

```

![Qitmeer Wallet](assets/wallet-info.png)



# RPC API

## account/key

```
listAccounts // list all keystore

createAccount // create a keystore (json key file and wallet database)

delAccount //del a keystore from keys dir and wallet database

importAccount //import keystore from json

exportAccount //export keysort json file


getBalance // get account balance or address balance

```

## address

```

listAddresses

createAddress 

```

## tx

```

getAllTx //get account or address all tx

```

## RawTx
> qitmeer rpc method

```
getRawTx

createRawTx

signRawTx

sendRawTx
```
