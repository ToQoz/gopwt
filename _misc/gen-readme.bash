#!/bin/bash

dummy_user=gopwter

root=$(cd $(dirname $0) && pwd)/..
out=$root/README.md

git_latest_tag=$(git describe --tags --abbrev=0)
cat $root/_misc/readme.header.md > $out
echo >> $out

# install
echo "## Install" >> $out
echo '
```
$ go get -u github.com/ToQoz/gopwt/...
```
' >> $out

echo "## Try" >> $out

echo '
go to project:

```
$ mkdir -p $GOPATH/src/github.com/$(whoami)/gopwtexample
$ cd $GOPATH/src/github.com/$(whoami)/gopwtexample
```

write main_test.go

```
$ cat <<EOF > main_test.go
package main

import (
	"flag"
	"github.com/ToQoz/gopwt"
	"github.com/ToQoz/gopwt/assert"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	flag.Parse()
	gopwt.Empower()
	os.Exit(m.Run())
}

func TestFoo(t *testing.T) {
	a := "xxx"
	b := "yyy"
	assert.OK(t, a == b, "a should equal to b")
}
EOF
```

run tests:

```
$ go test' >> $out

d=$(pwd)

ls $GOPATH/src/github.com/$dummy_user/gopwtexample || mkdir -p $GOPATH/src/github.com/$dummy_user/gopwtexample
cd $GOPATH/src/github.com/$dummy_user/gopwtexample

cat <<EOF > main_test.go
package main

import (
	"flag"
	"github.com/ToQoz/gopwt"
	"github.com/ToQoz/gopwt/assert"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	flag.Parse()
	gopwt.Empower()
	os.Exit(m.Run())
}

func TestFoo(t *testing.T) {
	a := "xxx"
	b := "yyy"
	assert.OK(t, a == b, "a should equal to b")
}
EOF

go test >> $out
echo '```' >> $out

cd $d

# example code
echo "" >> $out
echo "## Example" >> $out
echo '```go' >> $out
cat _example/main_test.go >> $out
echo '```' >> $out

# example result
echo "" >> $out
echo '```' >> $out
echo '$ go test' >> $out
cd _example
go test 2>&1 >> $out
echo '```' >> $out
echo "" >> $out

cat $root/_misc/readme.footer.md >> $out
