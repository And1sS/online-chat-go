#!/bin/sh

NODE_ID=$(head -c 12 /dev/random | base64)

mkdir -p /etc/consul
consul agent \
        -retry-join "$CONSUL_HOST" \
        -data-dir /data \
        -node redis-bus-"$NODE_ID" \
        -config-dir /etc/consul.d/client \
        --enable-local-script-checks > /etc/consul/consul.log &

redis-server