# Minimal SCION test network


## Install prerequisites

Reference platform: Ubuntu 22.04 LTS or macOS Version 14

Go 1.23: https://go.dev/dl/


## Set up SCION test environment

```
export SCION_PATH=$(pwd)/scion
export SCION_TESTNET_PATH=$(pwd)/scion-testnet

git clone https://github.com/scionproto/scion.git
git clone https://github.com/marcfrei/scion-testnet.git

cd $SCION_PATH
go build -o ./bin/ ./control/cmd/control
go build -o ./bin/ ./daemon/cmd/daemon
go build -o ./bin/ ./dispatcher/cmd/dispatcher
go build -o ./bin/ ./router/cmd/router
go build -o ./bin/ ./scion/cmd/scion
```

## Prepare SCION test network

```
cd $SCION_TESTNET_PATH
rm -rf logs gen-cache
mkdir logs gen-cache

go run scion-cryptogen.go topos/default
```

### On macOS:

```
for i in {2..255}; do sudo ifconfig lo0 alias 127.0.0.$i up; done
```

This will add alias IP addresses to the loopback adapter.

To later remove them again:

```
for i in {2..255}; do sudo ifconfig lo0 alias 127.0.0.$i down; done
for i in {2..255}; do sudo ifconfig lo0 -alias 127.0.0.$i; done
```


## Start SCION test network infrastructure

```
cd $SCION_TESTNET_PATH
sudo killall router control daemon dispatcher 2> /dev/null
./run-default.sh
```


## Use SCION test network

See [test topology](https://github.com/scionproto/scion/blob/master/doc/fig/default_topo.png)

```
$SCION_PATH/bin/scion --sciond 127.0.0.212:30255 address
$SCION_PATH/bin/scion --sciond 127.0.0.212:30255 showpaths -r --no-probe 1-ff00:0:133
$SCION_PATH/bin/scion --sciond 127.0.0.212:30255 ping --refresh -c 7 1-ff00:0:133,127.0.0.1
```


## Try test programs

In 1st session:

```
go run test-server.go -local 1-ff00:0:133,127.0.0.148:31000
```

In 2nd session:

```
go run test-client.go -daemon 127.0.0.212:30255 -local 2-ff00:0:222,127.0.0.212:31000 -remote 1-ff00:0:133,127.0.0.148:31000 -data "abc"
```
