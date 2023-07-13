#!/bin/sh

mkdir -p /etc/consul
consul agent \
        -retry-join "$CONSUL_HOST" \
        -data-dir /data \
        -node redis-bus \
        -config-dir /etc/consul.d/client \
        --enable-local-script-checks > /etc/consul/consul.log &

redis-server