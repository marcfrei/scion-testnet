# Minimal SCION test network


## Install prerequisites

Reference platform: Ubuntu 22.04 LTS or macOS Version 14

Go 1.23: https://go.dev/dl/

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
rm -rf logs gen-cache
mkdir logs gen-cache

go run scion-cryptogen.go topos/tiny4
```

### On macOS:

```
for i in {2..31}; do sudo ifconfig lo0 alias 127.0.0.$i up; done
```

This will add alias IP addresses to the loopback adapter.

To later remove them again:

```
for i in {2..31}; do sudo ifconfig lo0 alias 127.0.0.$i down; done
for i in {2..31}; do sudo ifconfig lo0 -alias 127.0.0.$i; done
```

### On Linux

```
for i in {2..31}; do sudo ifconfig lo:$i 127.0.0.$i up; done
```

This will add alias IP addresses to the loopback adapter (assuming your loopback adapter is called `lo`).

To later remove them again:

```
for i in {2..31}; do sudo ifconfig lo:$i 127.0.0.$i down; done
```


## Start SCION test network infrastructure

```
cd $SCION_TESTNET_PATH
sudo killall router control daemon dispatcher 2> /dev/null
./run-tiny4.sh
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
go run test-server.go -local 1-ff00:0:112,127.0.0.27:31000
```

In 2nd session:

```
go run test-client.go -daemon 127.0.0.19:30255 -local 1-ff00:0:111,127.0.0.19:31000 -remote 1-ff00:0:112,127.0.0.27:31000 -data "abc"
```


## Using a larger test topology

To use the larger [default test topology](https://github.com/scionproto/scion/blob/master/doc/fig/default_topo.png), apply the following changes:


### Prepare SCION test network

```
cd $SCION_TESTNET_PATH
rm -rf logs gen-cache
mkdir logs gen-cache

go run scion-cryptogen.go topos/default
```

#### On macOS:

Set up and tear down more loopback addresses.

```
for i in {2..255}; do sudo ifconfig lo0 alias 127.0.0.$i up; done
```

This will add alias IP addresses to the loopback adapter.

To later remove them again:

```
for i in {2..255}; do sudo ifconfig lo0 alias 127.0.0.$i down; done
for i in {2..255}; do sudo ifconfig lo0 -alias 127.0.0.$i; done
```
#### On Linux

```
for i in {2..255}; do sudo ifconfig lo:$i 127.0.0.$i up; done
```

This will add alias IP addresses to the loopback adapter (assuming your loopback adapter is called `lo`).

To later remove them again:

```
for i in {2..255}; do sudo ifconfig lo:$i 127.0.0.$i down; done
```

### Start SCION test network infrastructure

```
cd $SCION_TESTNET_PATH
sudo killall router control daemon dispatcher 2> /dev/null
./run-default.sh
```


### Use SCION test network

```
$SCION_PATH/bin/scion --sciond 127.0.0.212:30255 address
$SCION_PATH/bin/scion --sciond 127.0.0.212:30255 showpaths -r --no-probe 1-ff00:0:133
$SCION_PATH/bin/scion --sciond 127.0.0.212:30255 ping --refresh -c 7 1-ff00:0:133,127.0.0.1
```


### Try test programs

In 1st session:

```
go run test-server.go -local 1-ff00:0:133,127.0.0.148:31000
```

In 2nd session:

```
go run test-client.go -daemon 127.0.0.212:30255 -local 2-ff00:0:222,127.0.0.212:31000 -remote 1-ff00:0:133,127.0.0.148:31000 -data "abc"
```
