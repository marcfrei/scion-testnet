#!/usr/bin/env bash
set -Eeuo pipefail

$SCION_PATH/bin/router --config topos/tiny4/ASff00_0_110/br1-ff00_0_110-1.toml > logs/br1-ff00_0_110-1.log 2>&1 &
$SCION_PATH/bin/router --config topos/tiny4/ASff00_0_110/br1-ff00_0_110-2.toml > logs/br1-ff00_0_110-2.log 2>&1 &
$SCION_PATH/bin/control --config topos/tiny4/ASff00_0_110/cs1-ff00_0_110-1.toml > logs/cs1-ff00_0_110-1.log 2>&1 &
$SCION_PATH/bin/daemon --config topos/tiny4/ASff00_0_110/sd.toml > logs/sd1-ff00_0_110-1.log 2>&1 &
$SCION_PATH/bin/router --config topos/tiny4/ASff00_0_111/br1-ff00_0_111-1.toml > logs/br1-ff00_0_111-1.log 2>&1 &
$SCION_PATH/bin/control --config topos/tiny4/ASff00_0_111/cs1-ff00_0_111-1.toml > logs/cs1-ff00_0_111-1.log 2>&1 &
$SCION_PATH/bin/daemon --config topos/tiny4/ASff00_0_111/sd.toml > logs/sd1-ff00_0_111-1.log 2>&1 &
$SCION_PATH/bin/router --config topos/tiny4/ASff00_0_112/br1-ff00_0_112-1.toml > logs/br1-ff00_0_112-1.log 2>&1 &
$SCION_PATH/bin/control --config topos/tiny4/ASff00_0_112/cs1-ff00_0_112-1.toml > logs/cs1-ff00_0_112-1.log 2>&1 &
$SCION_PATH/bin/daemon --config topos/tiny4/ASff00_0_112/sd.toml > logs/sd1-ff00_0_112-1.log 2>&1 &
$SCION_PATH/bin/dispatcher --config topos/tiny4/dispatcher/disp.toml > logs/disp.log 2>&1 &
