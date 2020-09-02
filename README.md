# Go Blockchain Simulator

## Description

A simulated Blockchain written in Go. Followed [this article](https://hackernoon.com/learn-blockchains-by-building-one-117428612f46) for creating a single node, originally in python. Added some features to get it running as a network.

## Usage

This is a brief overview on how to get it running, to learn what it is actually doing [read this blog post.]()

1. First build the binary `cd go-blockchain && go build`

1. Open two terminals and start two nodes: `./go-blockchain 8000` `./go-blockchain 8001`

1. Register the nodes with each other by hitting the `/nodes/register` endpoint:
```
curl --location --request POST \
'http://localhost:8000/nodes/register' \
--header 'Content-Type: application/json' \
--data-raw '{
 "nodes": [
  "http://localhost:8001"
  ]
}'
```

1. Add transactions to the network:
```
curl --location --request POST \
'http://localhost:8000/transactions/new' \
--header 'Content-Type: application/json' \
--data-raw '{
 "sender":"sender1",
 "recipient":"recipient2",
 "amount": 12
}'
```

1. View the current chain before we mine a block:
```
curl http://localhost:8000/chain
```

1. Mine the block:
```
curl http://localhost:8000/mine
```

1. Find the newly replaced chain on the second node:
```
curl http://localhost:8001/chain
```

1. Create more transactions and mine more blocks!