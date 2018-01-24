package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"reflect"
	"sync"
	"testing"

	"github.com/ToQoz/gopwt"
	"github.com/ToQoz/gopwt/assert"
)

func TestMain(m *testing.M) {
	flag.Parse()
	gopwt.Empower() // magic happens
	os.Exit(m.Run())
}

func TestWithMessage(t *testing.T) {
	var receiver *struct{}
	receiver = nil
	assert.OK(t, receiver != nil, "receiver should not be nil")
}

func TestBasicLit(t *testing.T) {
	assert.OK(t, "a" == "b")
	assert.OK(t, 1 == 2)

	a := 1
	b := 2
	c := 3
	assert.OK(t, a+c == b)
	assert.OK(t, (a+c)+a == b)
	assert.OK(t, `foo
bar` == "bar")
}

func TestUnaryExpr(t *testing.T) {
	assert.OK(t, +1 == -1)

	a := "a"
	b := "a"
	assert.OK(t, &a == &b)

	wg := sync.WaitGroup{}
	ch := make(chan int)

	wg.Add(1)
	go func() {
		assert.OK(t, <-ch == 0)
		wg.Done()
	}()
	ch <- 1
	wg.Wait()
	close(ch)
}

func TestStringDiff(t *testing.T) {
	assert.OK(t, "supersoper" == "supersuper")

	expected := `foo
baz
bar2
bar`
	got := `foo
baz
bar
bar`
	assert.OK(t, got == expected)

	expected = `<div>
<div>
foo
</div>
</div>`
	got = `<div>
bar
</div>`
	assert.OK(t, got == expected)
}

func TestMapType(t *testing.T) {
	k := "b--------key"
	v := "b------value"
	assert.OK(t, reflect.DeepEqual(map[string]string{}, map[string]string{
		"a": "a",
		k:   v,
	}))
}

func TestArrayType(t *testing.T) {
	index := 1
	assert.OK(t, []int{
		1,
		2,
	}[index] == 3)
}

func TestStructType(t *testing.T) {
	foox := "foo------x"
	assert.OK(t, reflect.DeepEqual(struct{ Name string }{foox}, struct{ Name string }{"foo"}))
	assert.OK(t, reflect.DeepEqual(struct{ Name string }{Name: foox}, struct{ Name string }{Name: "foo"}))
}

func TestNestedCallExpr(t *testing.T) {
	rev := func(a bool) bool {
		return !a
	}

	assert.OK(t, rev(rev(rev(true))))
}

func TestCallWithNonIdempotentFunc(t *testing.T) {
	i := 0
	incl := func() int {
		i++
		return i
	}

	assert.OK(t, incl()+incl() == incl()+incl())
	assert.OK(t, incl() == incl())
	assert.OK(t, incl() == incl())
	assert.OK(t, incl() == incl())
	assert.OK(t, incl() == incl())
	assert.OK(t, (incl() == incl()) != (incl() == incl()))

	i2 := 0
	incl2 := func(i3 int) int {
		i2 += i3
		return i2
	}

	assert.OK(t, incl2(incl2(2)) == 10)
}

func TestPkgValue(t *testing.T) {
	assert.OK(t, sql.ErrNoRows == fmt.Errorf("error"))
}
