package translatedassert

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestReflectNilValue(t *testing.T) {
	type foo struct{}

	f := reflect.ValueOf(func(f *foo) {
		if f != nil {
			t.Error("translatedassert.reflectNilValue should return nil")
		}
	})
	f.Call([]reflect.Value{reflectNilValue(f.Type().In(0))})

	f = reflect.ValueOf(func(f interface{}) {
		if f != nil {
			t.Error("translatedassert.reflectNilValue should return nil")
		}
	})
	f.Call([]reflect.Value{reflectNilValue(f.Type().In(0))})

	f = reflect.ValueOf(func(f []string) {
		if f != nil {
			t.Error("translatedassert.reflectNilValue should return nil")
		}
	})
	f.Call([]reflect.Value{reflectNilValue(f.Type().In(0))})

	f = reflect.ValueOf(func(f map[string]string) {
		if f != nil {
			t.Error("translatedassert.reflectNilValue should return nil")
		}
	})
	f.Call([]reflect.Value{reflectNilValue(f.Type().In(0))})
}

func TestFormat(t *testing.T) {
	pvs := []posValuePair{
		NewPosValuePair(1, 2, true),
		NewPosValuePair(3, 2.5, true),
		NewPosValuePair(7, 5000, true),
	}
	e := format(pvs)
	o := trim(`
| |   |
| |   5000
| 2.5
2
`)

	if e != o {
		t.Error(fmt.Sprintf("(expected)%s != (got)%s\n", e, o))
	}
}

func TestFormat_String(t *testing.T) {
	pvs := []posValuePair{
		NewPosValuePair(1, "singleline", true),
		NewPosValuePair(5, "multi\nline\n-2", true),
		NewPosValuePair(9, "multi\nline", true),
	}
	e := format(pvs)
	o := trim(`
|   |   |
|   |   "multi
|   |   line"
|   "multi
|   line
|   -2"
"singleline"
`)

	if e != o {
		t.Error(fmt.Sprintf("(expected)\n%s", e))
		t.Error(fmt.Sprintf("(got)\n%s", o))
	}

	pvs = []posValuePair{
		NewPosValuePair(1, "a", true),
		NewPosValuePair(3, "tab\t", true),
		NewPosValuePair(5, "tab	", true),
		NewPosValuePair(10, "vtab\v", true),
	}
	e = format(pvs)
	o = trim(`
| | |    |
| | |    "vtab\v"
| | "tab\t"
| "tab\t"
"a"
`)
	if e != o {
		t.Error(fmt.Sprintf("(expected)%s != (got)%s\n", e, o))
	}
}

func TestTruncate(t *testing.T) {
	e := "to..."
	o := truncate("toqoz", 5, "...")
	if e != o {
		t.Error(fmt.Sprintf("(expected)%s != (got)%s\n", e, o))
	}

	e = " t..."
	o = truncate(" toqoz", 5, "...")
	if e != o {
		t.Error(fmt.Sprintf("(expected)%s != (got)%s\n", e, o))
	}
}

func trim(s string) string {
	return strings.TrimPrefix(s, "\n")
}
