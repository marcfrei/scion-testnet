# Minimal SCION test network


## Install prerequisites

Reference platform: Ubuntu 22.04 LTS or macOS Version 14

Go 1.22.2: https://go.dev/dl/


## Set up SCION test environment

```
export SCION_PATH=$(pwd)/scion
export SCION_TESTNET_PATH=$(pwd)/scion-testnet

git clone https://github.com/jordisubira/scion.git
git clone https://github.com/marcfrei/scion-testnet.git

cd $SCION_PATH
git checkout dispatcher_off
git apply $SCION_TESTNET_PATH/upstream.patch.0
git apply $SCION_TESTNET_PATH/upstream.patch.1
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

go run scion-cryptogen.go
```

### On macOS:

```
for i in {2..63}; do sudo ifconfig lo0 alias 127.0.0.$i up; done
```

This will add alias IP addresses to the loopback adapter.

To later remove them again:

```
for i in {2..63}; do sudo ifconfig lo0 alias 127.0.0.$i down; done
for i in {2..63}; do sudo ifconfig lo0 -alias 127.0.0.$i; done
```


## Start SCION test network infrastructure

```
cd $SCION_TESTNET_PATH
sudo killall router control daemon dispatcher 2> /dev/null
$SCION_PATH/bin/router --config gen/ASff00_0_110/br1-ff00_0_110-1.toml > logs/br1-ff00_0_110-1 2>&1 &
$SCION_PATH/bin/router --config gen/ASff00_0_110/br1-ff00_0_110-2.toml > logs/br1-ff00_0_110-2 2>&1 &
$SCION_PATH/bin/router --config gen/ASff00_0_110/br1-ff00_0_110-3.toml > logs/br1-ff00_0_110-3 2>&1 &
$SCION_PATH/bin/router --config gen/ASff00_0_110/br1-ff00_0_110-4.toml > logs/br1-ff00_0_110-4 2>&1 &
$SCION_PATH/bin/control --config gen/ASff00_0_110/cs1-ff00_0_110-1.toml > logs/cs1-ff00_0_110-1 2>&1 &
$SCION_PATH/bin/daemon --config gen/ASff00_0_110/sd.toml > logs/sd-ff00_0_110-1 2>&1 &
$SCION_PATH/bin/router --config gen/ASff00_0_111/br1-ff00_0_111-1.toml > logs/br1-ff00_0_111-1 2>&1 &
$SCION_PATH/bin/router --config gen/ASff00_0_111/br1-ff00_0_111-2.toml > logs/br1-ff00_0_111-2 2>&1 &
$SCION_PATH/bin/control --config gen/ASff00_0_111/cs1-ff00_0_111-1.toml > logs/cs1-ff00_0_111-1 2>&1 &
$SCION_PATH/bin/daemon --config gen/ASff00_0_111/sd.toml > logs/sd-ff00_0_111-1 2>&1 &
$SCION_PATH/bin/router --config gen/ASff00_0_112/br1-ff00_0_112-1.toml > logs/br1-ff00_0_112-1 2>&1 &
$SCION_PATH/bin/router --config gen/ASff00_0_112/br1-ff00_0_112-2.toml > logs/br1-ff00_0_112-2 2>&1 &
$SCION_PATH/bin/control --config gen/ASff00_0_112/cs1-ff00_0_112-1.toml > logs/cs1-ff00_0_112-1 2>&1 &
$SCION_PATH/bin/daemon --config gen/ASff00_0_112/sd.toml > logs/sd-ff00_0_112-1 2>&1 &
$SCION_PATH/bin/dispatcher --config gen/dispatcher/disp.toml > logs/disp 2>&1 &
```


## Use SCION test network

### In 1-ff00:0:111:

```
$SCION_PATH/bin/scion --sciond 127.0.0.36:30255 address
$SCION_PATH/bin/scion --sciond 127.0.0.36:30255 showpaths -r --no-probe 1-ff00:0:112
$SCION_PATH/bin/scion --sciond 127.0.0.36:30255 ping --refresh 1-ff00:0:112,127.0.0.43
```

### In 1-ff00:0:112:

```
$SCION_PATH/bin/scion --sciond 127.0.0.44:30255 address
$SCION_PATH/bin/scion --sciond 127.0.0.44:30255 showpaths -r --no-probe 1-ff00:0:111
$SCION_PATH/bin/scion --sciond 127.0.0.44:30255 ping --refresh 1-ff00:0:111,127.0.0.35
```
