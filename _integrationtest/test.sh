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
  rm -rf "$(dirname $workspace)"
  exit $retval
}
trap cleanup INT QUIT TERM EXIT

echo "workspace = $workspace"
cd "$workspace"

(
  echo "test dont_parse_testdata"
  cd "$workspace/dont_parse_testdata"
  go test
)

# Regression tests

for issue in issue33 issue40 issue44
do
  (
    echo "test $issue"
    cd "$workspace/regression/$issue"
    go test
    go test
    go test
  )
done

# test vendoring
(
  GOPATH=$(dirname $workspace)
  export GOPATH

  go get golang.org/x/tools/go/loader

  echo "test issue36"
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
