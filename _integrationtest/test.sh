#!/bin/bash

set -ex

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
)

(
  cd "$workspace/regression/issue36"
  go get -v github.com/bmuschko/go-testing-frameworks/calc
  if [ ! -e "$GOPATH/bin/dep" ] && [ ! -e "$workspace/dep"  ]; then
    curl https://raw.githubusercontent.com/golang/dep/master/install.sh | INSTALL_DIRECTORY="$workspace" sh
  fi
  "$workspace/dep" ensure
  go test
)
