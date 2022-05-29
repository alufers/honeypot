#!/bin/bash
set -e


HOSTNAME="huawei-2137"

HARDENING="
    --memory 64m \
    --memory-swap=64m \
    --kernel-memory=64m \
    --cpus=0.25 \
    --network=none \
    --pids-limit=100 \
"

MIMICKING="
    --hostname=$HOSTNAME \
"

docker build -t honeypot-sandbox .

docker run \
    $HARDENING \
    $MIMICKING \
    --interactive \
    --tty \
    --rm \
    honeypot-sandbox /bin/sh
