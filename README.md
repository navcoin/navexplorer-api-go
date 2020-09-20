# navexplorer-api-go
REST API for Navexplorer.com

## To run locally

```
go get -d ./...
go run main.go
```

## Available endpoints

```
GET    /address?size=100
GET    /address/:hash
GET    /address/:hash/summary
GET    /address/:hash/history
GET    /address/:hash/staking

GET    /address/:hash/assoc/staking
GET    /balance
GET    /bestblock
GET    /blockcycle
GET    /blockgroup

GET    /block
GET    /block/:hash
GET    /block/:hash/cycle
GET    /block/:hash/raw
GET    /block/:hash/tx

GET    /tx/:hash
GET    /tx/:hash/raw

GET    /staking/blocks
GET    /staking/rewards

GET    /softfork
GET    /softfork/cycle

GET    /dao/consensus/parameters
GET    /dao/consensus/parameters/:id
GET    /dao/consultation
GET    /dao/consultation/:hash
GET    /dao/consultation/:hash/:answer/votes
GET    /dao/answer/:hash

GET    /dao/cfund/stats
GET    /dao/cfund/proposal
GET    /dao/cfund/proposal/:hash
GET    /dao/cfund/proposal/:hash/votes
GET    /dao/cfund/proposal/:hash/trend
GET    /dao/cfund/proposal/:hash/payment-request
GET    /dao/cfund/payment-request
GET    /dao/cfund/payment-request/:hash
GET    /dao/cfund/payment-request/:hash/votes
GET    /dao/cfund/payment-request/:hash/trend

GET    /search
```

## Network Header

Use the Network header to switch between the available NavCoin networks.

Set `header('Network: mainnet')` for mainnet data

Set `header('Network: testnet')` for testnet data
