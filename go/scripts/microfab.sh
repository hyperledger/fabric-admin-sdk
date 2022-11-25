#!/usr/bin/env bash

set -e

fail() {
    echo "$1"
    exit 1
}

isRunning() {
    docker inspect microfab &>/dev/null
}

up() {
    if ! isRunning; then
        if [ -n "$1" ]; then
            MICROFAB_CONFIG=$(< "$1") docker run --name microfab --rm --detach --publish 8080:8080 --env MICROFAB_CONFIG ibmcom/ibp-microfab
        else
            docker run --name microfab --rm --detach --publish 8080:8080 ibmcom/ibp-microfab
        fi
    fi

    # Wait for peers to join channel
    grep --max-count=1 'Committed block \[1\]' <(docker logs --follow microfab 2>&1) >/dev/null || fail 'ERROR: Microfab failed to start'

    
}

down() {
    if isRunning; then
        docker stop microfab
    fi
}

case "$1" in
up)
    up "$2"
    ;;
down)
    down
    ;;
*)
    SCRIPT_NAME="${BASH_SOURCE[0]##*/}"
    echo "Usage: ${SCRIPT_NAME} up [configfile] | down"
    exit 1
    ;;
esac
