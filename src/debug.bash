#!/usr/bin/env bash

ulimit -c unlimited

# Linux
# echo "/workdir/go/core.%e.%p" | sudo tee /proc/sys/kernel/core_pattern

# NetBSD
sysctl -w proc.$$.corename=$(dirname $0)/%n.%p.core

export GOTRACEBACK=crash
$(dirname $0)/all.bash
