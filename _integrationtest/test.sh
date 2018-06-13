#!/bin/bash

set -e

ORIGINAL_GOPATH=$GOPATH

# dep is not working with  _ prefiexed ( like a `_integrationtest` ) dir
workspace="/tmp/gopath-gopwt-integrationtest-$$/src"
mkdir -p "$workspace"
cp -r "$(cd "$(dirname "$0")"; pwd)"/* "$workspace/"

GOPATH=$(dirname $workspace):$GOPATH
export GOPATH

cleanup() {
  retval=$?
  #rm -rf "$workspace"
  exit $retval
}
trap cleanup INT QUIT TERM EXIT

echo "workspace = $workspace"
cd "$workspace"

(
  cd "$workspace/regression/issue33"
  go test
  go test
  go test
)

(
  cd "$workspace/regression/issue40"
  go test
  go test
  go test
)

(
  GOPATH=$(dirname $workspace)
  export GOPATH

  cd "$workspace/regression/issue36"
  dep ensure
  rm -rf "vendor/github.com/ToQoz/gopwt"
  cp -r "$ORIGINAL_GOPATH/src/github.com/ToQoz/gopwt" "vendor/github.com/ToQoz/"

  # for go1.8-
  find vendor -type f | grep '_test.go$' | xargs rm
  go test ./...
  go test ./...
  go test ./...
)
