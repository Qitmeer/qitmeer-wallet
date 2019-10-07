## qitmeer-wallet
The command-line wallet of the Qitmeer network


## download or build

you can download from [release](https://github.com/Qitmeer/qitmeer-wallet/releases)

if you have golang environment, you can build it by yourself

```

git clone https://github.com/Qitmeer/qitmeer-wallet.git

go build


```

## usage

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

## console model
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
            <getlisttxbyaddr> : Gets all transaction records at the specified address. Parameter: [address]
            <getNewAddress> : Create a new address under the account. Parameter: [account]
            <getAddressesByAccount> : Check all addresses under the account. Parameter: [account]
            <getAccountByAddress> : Inquire about the account number of the address. Parameter: [address]
            <importPrivKey> : Import private key. Parameter: [prikey]
            <importWifPrivKey> : Import wif format private key. Parameter: [prikey]
            <dumpPrivKey> : Export wif format private key by address. Parameter: [address]
            <getAccountAndAddress> : Check all accounts and addresses. Parameter: []
            <sendToAddress> : Transfer transaction. Parameter: [address] [num]
            <updateblock> : Update Wallet Block. Parameter: []
            <syncheight> : Current Synchronized Data Height. Parameter: []
            <unlock> : Unlock Wallet. Parameter: [password]
            <help> : help
            <exit> : Exit command mode

```


## How to use qitmeer-wallet console command model

- create walllet

    `
        ./qitmeer-wallet -console
    `
    
    `#output`
    
    ```
    Config.Cfg.AppDataDir:/root/.qitwallet
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
    
    
- Import private key

    `[wallet-cli]: importPrivKey 123456 `
    
    `#output`
    
    `ImportPrivKey: OK`
    
- Create a new account
    
    `[wallet-cli]: createNewAcceount test`
    
    `#output`
    
    `CreateNewAccount: succ`

- Check all your account addresses
  
  `[wallet-cli]: getAccountAndAddress`
  
  `#output`
  
  `Account: imported, address: TmgD1mu8zMMV9aWmJrXqYnWRhR9SBfDZG6
   Account: imported, address: TmK8tyqW9hvoT1J1qXRzU8C4m6fZ6zigD4
   Account: imported, address: Tmbsds jwzuGboFQ9GcKg6EUmrr3tokzozyF
   `

- View address balance

    `[wallet-cli]: getbalance TmbsdsjwzuGboFQ9GcKg6EUmrr3tokzozyF`
    
    `#output`
    
    `Getbalance amount: 0.04497 MEER`
    
- Unlock Wallet Auto-lock after 15 minutes
    
    `[wallet-cli]: unlock 123456`
    
    `#output`
    
    `Unlock succ`
    
- Synchronized blocks are automatically updated by default and checked once a minute

    `[wallet-cli]: updateblock`
    
    `#output`
    
    `updpateblock start`
    
- Transfer from a wallet requires unlocking the wallet first

    `[wallet-cli]: sendToAddress TmbsdsjwzuGboFQ9GcKg6EUmrr3tokzozyF 0.01`
    
    `#output`
    
    `0e441ecee44defe28711103eef0cc3d01c187c257738150869c032fbbf96d4c9`
    
- View the current block synchronization number

    `[wallet-cli]: syncheight`
    
    `#output`
    
    `5433`
    

- View all transaction records corresponding to address

    `[wallet-cli]: getlistTXbyaddr TmbsdsjwzuGboFQ9GcKg6EUmrr3tokzozyF`
  
- View all transaction records corresponding to address

    `[wallet-cli]: getlistTXbyaddr TmbsdsjwzuGboFQ9GcKg6EUmrr3tokzozyF`
    

  
- exit

    `[wallet-cli]: exit`
    
