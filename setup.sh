#!/usr/bin/env bash
set -o nounset
set -o errexit
set -o xtrace

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"

# grab dependencies
builtin cd "$DIR"
git submodule update --init --recursive

# build and run
export GOPATH="$DIR"
builtin cd "$DIR/src/app"
go build
./app
