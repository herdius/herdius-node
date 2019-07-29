# herdius-node
Herdius validator node

## Build and Run

```
$ make start-node PORT=3001 PEERS="tcp://127.0.0.1:3000" SELFIP="127.0.0.1"
```
or

```
make build-node
./node -selfip 127.0.0.1 -port 3001 -peers tcp://127.0.0.1:3000 

```
