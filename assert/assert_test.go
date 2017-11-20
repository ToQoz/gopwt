package assert

import (
	"reflect"
	"testing"
)

type testingDummy struct {
	errorCalled     bool
	errorCalledWith []interface{}
	skipCalled      bool
	skipCalledWith  []interface{}
}

func (t *testingDummy) Error(args ...interface{}) {
	t.errorCalled = true
	t.errorCalledWith = args
}

func (t *testingDummy) Skip(args ...interface{}) {
	t.skipCalled = true
	t.skipCalledWith = args
}

func TestOK(t *testing.T) {
	td := &testingDummy{}

	_ok(td, true, "caller line")
	if td.errorCalled {
		t.Error("t.Error should not be called if ok")
	}

	_ok(td, false, "caller line")
	if !td.errorCalled {
		t.Error("t.Error should be called if not ok")
	}
	errorArgs := []interface{}{`[FAIL Assersion] caller line

Please call gopwt.Empower() in your TestMain(t *testing.M). It give you power.
If you need more information, see http://github.com/ToQoz/gopwt
`}
	if !reflect.DeepEqual(td.errorCalledWith, errorArgs) {

		t.Errorf("t.Error should be called with args.\n(expected)=%+v\n(got)=%+v\n", errorArgs, td.errorCalledWith)
	}
}

func TestOKWithAssersionMessage(t *testing.T) {
	td := &testingDummy{}

	_ok(td, false, "caller line", "dummy message")
	if !td.errorCalled {
		t.Error("t.Error should be called if not ok")
	}
	expected := []interface{}{`[FAIL Assersion] caller line

Please call gopwt.Empower() in your TestMain(t *testing.M). It give you power.
If you need more information, see http://github.com/ToQoz/gopwt

AssersionMessage:
	- dummy message`}
	if !reflect.DeepEqual(td.errorCalledWith, expected) {

		t.Errorf("t.Error should be called with args.\n(expected)=%+v\n(got)=%+v\n", expected, td.errorCalledWith)
	}
}

func TestRequired(t *testing.T) {
	td := &testingDummy{}

	_require(td, true, "caller line")
	if td.errorCalled {
		t.Error("t.Error should not be called if ok")
	}
	if td.skipCalled {
		t.Error("t.Skip should not be called if ok")
	}

	_require(td, false, "caller line")
	if !td.errorCalled {
		t.Error("t.Error should be called if not ok")
	}
	expected := []interface{}{`[FAIL Assersion] caller line

Please call gopwt.Empower() in your TestMain(t *testing.M). It give you power.
If you need more information, see http://github.com/ToQoz/gopwt
`}
	if !reflect.DeepEqual(td.errorCalledWith, expected) {

		t.Errorf("t.Error should be called with args.\n(expected)=%+v\n(got)=%+v\n", expected, td.errorCalledWith)
	}

	if !td.skipCalled {
		t.Error("t.Skip should be called if not ok")
	}
	skipArgs := []interface{}{"skip by gopwt/assert.Require"}
	if !reflect.DeepEqual(td.skipCalledWith, skipArgs) {

		t.Errorf("t.Skip should be called with args.\n(expected)=%+v\n(got)=%+v\n", skipArgs, td.skipCalledWith)
	}
}

func TestRequiredWithAssersionMessage(t *testing.T) {
	td := &testingDummy{}

	_require(td, false, "caller line", "dummy message")
	if !td.errorCalled {
		t.Error("t.Error should be called if not ok")
	}
	errorArgs := []interface{}{`[FAIL Assersion] caller line

Please call gopwt.Empower() in your TestMain(t *testing.M). It give you power.
If you need more information, see http://github.com/ToQoz/gopwt

AssersionMessage:
	- dummy message`}

	if !reflect.DeepEqual(td.errorCalledWith, errorArgs) {

		t.Errorf("t.Error should be called with args.\n(expected)=%+v\n(got)=%+v\n", errorArgs, td.errorCalledWith)
	}

	if !td.skipCalled {
		t.Error("t.Skip should be called if not ok")
	}
	skipArgs := []interface{}{"skip by gopwt/assert.Require"}
	if !reflect.DeepEqual(td.skipCalledWith, skipArgs) {

		t.Errorf("t.Skip should be called with args.\n(expected)=%+v\n(got)=%+v\n", skipArgs, td.skipCalledWith)
	}
}

func TestCallerLine(t *testing.T) {
	var l string

	a := func() {
		l = callerLine(1)
	}
	a()
	if l != "a()" {
		t.Errorf("expected %s but got %s", "a()", l)
	}

	l = callerLine(0)
	if l != "l = callerLine(0)" {
		t.Errorf("expected %s but got %s", "l = callerLine(0)", l)
	}

	b := func() {
		l = callerLine(0)
	}
	b()
	if l != "l = callerLine(0)" {
		t.Errorf("expected %s but got %s", "l = callerLine(0)", l)
	}
}
