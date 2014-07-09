package translatedassert

import (
	"fmt"
	"strings"
	"testing"
)

func TestFormat(t *testing.T) {
	pvs := []posValuePair{
		NewPosValuePair(1, 2),
		NewPosValuePair(3, 2.5),
		NewPosValuePair(7, 5000),
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
		NewPosValuePair(1, "singleline"),
		NewPosValuePair(5, "multi\nline\n-2"),
		NewPosValuePair(9, "multi\nline"),
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
		t.Error(fmt.Sprintf("(expected)%s != (got)%s\n", e, o))
	}

	pvs = []posValuePair{
		NewPosValuePair(1, "a"),
		NewPosValuePair(3, "tab\t"),
		NewPosValuePair(5, "tab	"),
		NewPosValuePair(10, "vtab\v"),
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
