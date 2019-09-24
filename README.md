# qitmeer-wallet
The command-line wallet of the Qitmeer network


# download or build

you can download from [release](https://github.com/Qitmeer/qitmeer-wallet/releases)

if you have golang environment, you can build it by yourself

```

git clone https://github.com/Qitmeer/qitmeer-wallet.git

go build


```

# usage

There are two ways to use it：RPC interface and Web interface.


## RPC interface

```
# start qitmeer-wallet

./qitmeer-wallet -n testnet --listens=127.0.0.1:38130 --rpcUser admin --rpcPass 123

# default RPC port is 38130

# example

#curl

curl -i -X POST -H 'Content-type':'application/json' --user uid:pwd -d '{"jsonrpc": "2.0","method": "getBalance","params": ["your-address"],"id": 1}' http://127.0.0.1:38130

```

you can also use [qitmeer-cli](https://github.com/Qitmeer/qitmeer-cli) to access the qitmeer-wallet RPC interface.

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

### console model
```
    # start qitmerr-wallet 
    ./qitmeer-wallet -console
    
    # get helper
    [wallet-cli]: help
    Usage:
            <command> [arguments]
            The commands are:
            <createNewAccount> : Create a new account. Parameter: [account]
            <getbalance> : Query the specified address balance. Parameter: [address]
            <listAccountsBalance> : Obtain all account balances. Parameter: []
            <getlisttxbyaddr> : Gets all transaction records at the specified address. Parameter: [address]
            <getNewAddress> : Create a new address under the account. Parameter: [account]
            <getAddressesByAccount> : Check all addresses under the account. Parameter: [account]
            <getAccountByAddress> : Inquire about the account number of the address. Parameter: [address]
            <importPrivKey> : Import private key. Parameter: [prikey]
            <importWifPrivKey> : Import wif format private key. Parameter: [prikey]
            <dumpPrivKey> : Export the private key by address. Parameter: [address]
            <getAccountAndAddress> : Check all accounts and addresses. Parameter: []
            <sendToAddress> : Transfer transaction. Parameter: [address] [num]
            <updateblock> : Update Wallet Block. Parameter: []
            <syncheight> : Current Synchronized Data Height. Parameter: []
            <help> : help
            <exit> : Exit command mode

    [wallet-cli]: createNewAccount test
```


