package assert

import (
	"io/ioutil"
	"runtime"
	"strings"
	"testing"

	_ "github.com/ToQoz/gopwt/translatedassert" // for go get github.com/ToQoz/gopwt/assert
)

type testingInterface interface {
	Error(args ...interface{})
	Skip(args ...interface{})
}

// OK assert given bool is true.
// This will be translate to `translatedassert.OK` by gopwt.
func OK(t *testing.T, ok bool, messages ...string) {
	_ok(t, ok, callerLine(1), messages...)
}

func _ok(t testingInterface, ok bool, callerLine string, messages ...string) {
	if ok {
		return
	}

	msg := `[FAIL Assersion] ` + callerLine + `

Please run tests by command "gopwt". It give you power.
If you need more information, see http://github.com/ToQoz/gopwt
`

	if len(messages) > 0 {
		msg += "\nAssersionMessage:\n"
		for _, m := range messages {
			msg += "\t- " + m
		}
	}

	t.Error(msg)
}

// Require assert given bool is true.
// The difference from OK is Require calls t.Skip after t.Error on fail.
// This will be translate to `translatedassert.Require` by gopwt.
func Require(t *testing.T, ok bool, messages ...string) {
	_require(t, ok, callerLine(1), messages...)
}

func _require(t testingInterface, ok bool, callerLine string, messages ...string) {
	if ok {
		return
	}

	msg := `[FAIL Assersion] ` + callerLine + `

Please run tests by command "gopwt". It give you power.
If you need more information, see http://github.com/ToQoz/gopwt
`

	if len(messages) > 0 {
		msg += "\nAssersionMessage:\n"
		for _, m := range messages {
			msg += "\t- " + m
		}
	}

	t.Error(msg)
	t.Skip("skip by gopwt/assert.Require")
}

func callerLine(skip int) string {
	_, file, lnum, ok := runtime.Caller(skip + 1)

	if !ok {
		return ""
	}

	b, err := ioutil.ReadFile(file)
	if err != nil {
		return ""
	}

	s := strings.Split(string(b), "\n")
	if len(s) <= lnum-1 {
		return ""
	}

	return strings.Trim(s[lnum-1], " \n\t")
}
