# Minimal SCION test network


## Install prerequisites

Reference platform: Ubuntu 22.04 LTS or macOS Version 14

Go 1.24: https://go.dev/dl/

### Windows

Currently SCION is not supported on Windows, hence using WSL is recommended.

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
sudo go run scion-testnet.go ifconfig topos/tiny4
go run scion-testnet.go cryptogen topos/tiny4
```


## Start SCION test network infrastructure

```
go run scion-testnet.go run topos/tiny4
```


## Use SCION test network

```
$SCION_PATH/bin/scion --sciond 127.0.0.19:30255 address
$SCION_PATH/bin/scion --sciond 127.0.0.19:30255 showpaths -r --no-probe 1-ff00:0:112
$SCION_PATH/bin/scion --sciond 127.0.0.19:30255 ping --refresh -c 7 1-ff00:0:112,127.0.0.1
```


## Try test programs

In 1st session:

```
go run test-server.go -local 1-ff00:0:112,127.0.0.28:31000
```

In 2nd session:

```
go run test-client.go -daemon 127.0.0.19:30255 -local 1-ff00:0:111,127.0.0.20:31000 -remote 1-ff00:0:112,127.0.0.28:31000 -data "abc"
```


## Remove SCION test network

```
Terminate test network with Ctrl+C

sudo go run scion-testnet.go ifconfig -c topos/tiny4
```


## Using a larger test topology

To use the larger [default test topology](https://github.com/scionproto/scion/blob/master/doc/fig/default_topo.png), apply the following changes:


### Prepare SCION test network

```
sudo go run scion-testnet.go ifconfig topos/default
go run scion-testnet.go cryptogen topos/default
```


### Start SCION test network infrastructure

```
go run scion-testnet.go run topos/default
```


### Use SCION test network

```
$SCION_PATH/bin/scion --sciond '[fd00:f00d:cafe::7f00:54]:30255' address
$SCION_PATH/bin/scion --sciond '[fd00:f00d:cafe::7f00:54]:30255' showpaths -r --no-probe 1-ff00:0:133
$SCION_PATH/bin/scion --sciond '[fd00:f00d:cafe::7f00:54]:30255' ping --refresh -c 7 1-ff00:0:133,127.0.0.1
```


### Try test programs

In 1st session:

```
go run test-server.go -local 1-ff00:0:133,127.0.0.101:31000
```

In 2nd session:

```
go run test-client.go -daemon '[fd00:f00d:cafe::7f00:54]:30255' -local '2-ff00:0:222,[fd00:f00d:cafe::7f00:55]:31000' -remote 1-ff00:0:133,127.0.0.101:31000 -data "abc"
```


### Remove SCION test network

```
Terminate test network with Ctrl+C

sudo go run scion-testnet.go ifconfig -c topos/default
```
