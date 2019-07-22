# herdius-node
Herdius validator node

## Enable Go module

Enable go module:

```
export GO111MODULE=on
```

## Build and Run

```
$ cd cmd/validator
$ go build
$ ./validator -peers='tcp://127.0.0.1:3000' -env=dev -port=3001
```
