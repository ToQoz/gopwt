// Package translatedassert is not for human.
package translatedassert

import (
	"fmt"
	"github.com/mattn/go-runewidth"
	"reflect"
	"strings"
	"testing"
)

var cachedFuncRet = map[string]map[int]interface{}{}

type posValuePair struct {
	Pos   int
	Value interface{}
}

// NewPosValuePair has nodoc
// **This is not for human**
func NewPosValuePair(pos int, v interface{}) posValuePair {
	return posValuePair{Pos: pos, Value: v}
}

// OK has nodoc
// **This is not for human**
func OK(t *testing.T, e bool, header, filename string, line int, origexpr string, termw int, pvPairs ...posValuePair) {
	if e {
		return
	}

	output := fmt.Sprintf("%s %s:%d\n%s", header, filename, line, origexpr) + "\n" + format(pvPairs)

	lines := []string{}

	for _, line := range strings.Split(output, "\n") {
		tabw := 8 // FIXME

		// $GOROOT/src/pkg/testing/testing.go
		// 194 func decorate(s string) string {
		// 217 // Second and subsequent lines are indented an extra tab.
		// 218 buf.WriteString("\n\t\t")
		terrorw := tabw * 2

		if termw > 0 && runewidth.StringWidth(line) > termw-terrorw {
			lines = append(lines, truncate(line, termw-terrorw, "[ommitted]..."))
		} else {
			lines = append(lines, line)
		}
	}

	t.Error(strings.Join(lines, "\n"))
	deleteCachedFuncRetsForFileline(filename, line)
}

// --------------------------------------------------------------------------------
// Reflect
// --------------------------------------------------------------------------------

// MFCall momorize and call func
func MFCall(filename string, line, col int, f reflect.Value, args ...reflect.Value) []reflect.Value {
	rets := getCachedFuncRetsForFileline(filename, line)

	if _, ok := rets[col]; !ok {
		for i, a := range args {
			// hack to passing nil...
			if !a.IsValid() {
				args[i] = reflectNilValue(f.Type().In(i))
			}
		}

		ret := FRVInterface(f.Call(args))
		rets[col] = ret
	}

	return []reflect.Value{reflect.ValueOf(rets[col])}
}

func getCachedFuncRetsForFileline(filename string, line int) map[int]interface{} {
	key := fmt.Sprintf("%s:%d", filename, line)

	if _, ok := cachedFuncRet[key]; !ok {
		cachedFuncRet[key] = map[int]interface{}{}
	}

	return cachedFuncRet[key]
}

func deleteCachedFuncRetsForFileline(filename string, line int) {
	key := fmt.Sprintf("%s:%d", filename, line)
	delete(cachedFuncRet, key)
}

// RVOf is reflect.ValueOf
func RVOf(a interface{}) reflect.Value {
	return reflect.ValueOf(a)
}

// RTOf is reflect.TypeOf
func RTOf(a interface{}) reflect.Type {
	return reflect.TypeOf(a)
}

// RVBool returns reflect value as bool
func RVBool(rv reflect.Value) bool {
	return rv.Bool()
}

// FRVBool returns first of reflect values as bool
func FRVBool(rvs []reflect.Value) bool {
	return RVBool(rvs[0])
}

// RVInterface returns reflect value as interface{}
func RVInterface(rv reflect.Value) interface{} {
	if !rv.IsValid() {
		return nil
	}

	return rv.Interface()
}

// FRVInterface returns first of reflect values as interface{}
func FRVInterface(rvs []reflect.Value) interface{} {
	rv := rvs[0]

	return RVInterface(rv)
}

func reflectNilValue(t reflect.Type) reflect.Value {
	return reflect.New(t).Elem()
}

// --------------------------------------------------------------------------------
// Format
// --------------------------------------------------------------------------------

func format(pvPairs []posValuePair) string {
	buf := make([]byte, 0, 50*len(pvPairs)*len(pvPairs)) // Try to avoid more allocations

	// first line
	cx := 0
	for x := 0; x < len(pvPairs); x++ {
		for i := 0; i < pvPairs[x].Pos-cx-1; i++ {
			buf = append(buf, byte(' '))
		}

		cx = pvPairs[x].Pos - 1
		buf = append(buf, byte('|'))
		cx++
	}
	buf = append(buf, byte('\n'))

	for y := 0; y < len(pvPairs); y++ {
		cx := 0

		for x := 0; x < len(pvPairs); x++ {
			if len(pvPairs)-y-1 != x {
				continue
			}

			// print "foo\nhoge" as "foo
			//                      hoge"
			for lnum, lstr := range strings.Split(formatValue(pvPairs[x].Value), "\\n") {
				if lnum > 0 {
					buf = append(buf, byte('\n'))
					cx = 0
				}

				// fill by space or vbar
				// e.g.
				// |  |
				// |  foo
				//[| ]bar
				for i := 0; i < pvPairs[x].Pos-1; i++ {
					found := false

					for _, n := range pvPairs[:x] {
						if n.Pos != i+1 {
							continue
						}

						found = true
						buf = append(buf, byte('|'))
						cx++
						break
					}

					if !found {
						buf = append(buf, byte(' '))
						cx++
					}
				}

				// embed value
				// e.g.
				// |  |
				// |  foo
				// | [bar]
				buf = append(buf, lstr...)
			}

			break
		}

		buf = append(buf, byte('\n'))
	}

	return string(buf)
}

// TODO use own impl instead of fmt.Sprintf("%#v", a)
func formatValue(a interface{}) string {
	return fmt.Sprintf("%#v", a)
}

func truncate(s string, w int, tail string) string {
	if w-len(tail) <= 0 {
		return tail
	}

	return _truncate(s, w-len(tail)) + tail
}

func _truncate(s string, w int) string {
	rs := []rune(s)
	i := 0

	cw := 0

	for _, r := range rs {
		rw := runewidth.RuneWidth(r)
		if cw+rw <= w {
			i++
			cw += rw
		} else {
			break
		}
	}

	return string(rs[0:i])
}
