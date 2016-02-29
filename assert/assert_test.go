package assert

import (
	"testing"
)

type testingDummy struct {
	errorCalled bool
	skipCalled  bool
}

func (t *testingDummy) Error(args ...interface{}) {
	t.errorCalled = true
}

func (t *testingDummy) Skip(args ...interface{}) {
	t.skipCalled = true
}

func TestOK(t *testing.T) {
	td := &testingDummy{}

	_ok(t, true, "caller line")
	if td.errorCalled {
		t.Error("t.Error should not be called if ok")
	}

	_ok(td, false, "caller line")
	if !td.errorCalled {
		t.Error("t.Error should be called if not ok")
	}
}

func TestRequired(t *testing.T) {
	td := &testingDummy{}

	_require(t, true, "caller line")
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
	if !td.skipCalled {
		t.Error("t.Skip should be called if not ok")
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
