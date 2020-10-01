#!/usr/bin/env bash
set -o nounset
set -o errexit
set -o xtrace

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"

git submodule update --init --recursive
export GOPATH="$DIR"
builtin cd "$DIR/src/app"
go build
./app
