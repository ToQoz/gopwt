#!/bin/bash

set -e

export GOPWT_DEBUG=1
ORIGINAL_GOPATH=${GOPATH:-$HOME/go}

# dep is not working with  _ prefiexed ( like a `_integrationtest` ) dir
workspace="/tmp/gopath-gopwt-integrationtest-$$/src"
mkdir -p "$workspace"
cp -r "$(cd "$(dirname "$0")"; pwd)"/* "$workspace/"

GOPATH=$(dirname $workspace):$GOPATH
export GOPATH

cleanup() {
  retval=$?
  rm -rf "$(dirname $workspace)"
  exit $retval
}
trap cleanup INT QUIT TERM EXIT

go_test() {
  echo "> go test $*"
  go test "$@"
}

echo "workspace = $workspace"
cd "$workspace"

(
  echo
  echo "[test dont_parse_testdata]"
  cd "$workspace/dont_parse_testdata"
  go_test
)

# Regression tests

for issue in issue33 issue40 issue44
do
  (
    echo
    echo "[test $issue]"
    rm -rf "$HOME/.gopwtcache"
    cd "$workspace/regression/$issue"
    go_test
    go_test
    go_test
  )
done

# test vendoring
(
  GOPATH=$(dirname $workspace)
  export GOPATH

  echo
  echo "[test issue36]"
  rm -rf "$HOME/.gopwtcache"
  cd "$workspace/regression/issue36"
  dep ensure -v -vendor-only
  rm -rf "vendor/github.com/ToQoz/gopwt"
  cp -r "$ORIGINAL_GOPATH/src/github.com/ToQoz/gopwt" "vendor/github.com/ToQoz/"

  # for go1.8-
  find vendor -type f | grep '_test.go$' | xargs rm
  go_test -v ./...
  go_test -v ./...
  go_test -v ./...
)
