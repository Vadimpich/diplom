#!/usr/bin/env sh
set -eu

export LD_LIBRARY_PATH="${LD_LIBRARY_PATH:+$LD_LIBRARY_PATH:}/usr/local/bin/WiMi/libs"
export LD_PRELOAD="${LD_PRELOAD:+$LD_PRELOAD:}/usr/local/bin/WiMi/libs/libjemalloc.so.2"

cd /usr/local/bin/WiMi
exec ./WiMi -e
