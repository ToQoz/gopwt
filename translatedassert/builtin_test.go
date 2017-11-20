package translatedassert

import (
	"reflect"
	"testing"
)

func TestBappend(t *testing.T) {
	b := []string{"1"}
	got := Bappend(b, "2")
	expected := []string{"1", "2"}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected=%#v, got=%#v", expected, got)
	}
}

func TestBcap(t *testing.T) {
	a := []string{"1"}
	got := Bcap(a)
	expected := cap(a)
	if got != expected {
		t.Errorf("Bcap(slice) should return capacity of slice. expected=%d, got=%d", expected, got)
	}

	b := [2]string{"1"}
	got = Bcap(b)
	expected = 2
	if got != expected {
		t.Errorf("Bcap(array) should return capacity of array. expected=%d, got=%d", expected, got)
	}
}

func TestBcomplex(t *testing.T) {
	if complex(1, 2) != Bcomplex(1, 2) {
		t.Errorf("Bcomplex behavior should be equal to complex")
	}

	if complex(1.0, 2.2) != Bcomplex(1.0, 2.2) {
		t.Errorf("Bcomplex behavior should be equal to complex, %#v %#v", complex(1.0, 2.2), Bcomplex(1.0, 2.2))
	}

	if complex(float64(1.0), float64(2.2)) != Bcomplex(float64(1.0), float64(2.2)) {
		t.Errorf("Bcomplex behavior should be equal to complex, %#v %#v", complex(1.0, 2.2), Bcomplex(1.0, 2.2))
	}

	if complex(float32(1.0), float32(2.2)) != Bcomplex(float32(1.0), float32(2.2)).(complex64) {
		t.Errorf("Bcomplex behavior should be equal to complex, %#v %#v", complex(1.0, 2.2), Bcomplex(1.0, 2.2))
	}
}

func TestBcopy(t *testing.T) {
	src := []int{1, 2, 3}
	dst1 := make([]int, 3)
	dst2 := make([]int, 3)
	Bcopy(dst1, src)
	copy(dst2, src)
	if !reflect.DeepEqual(dst1, dst2) {
		t.Errorf("Bcopy behavior should be equal to copy")
	}

}

func TestBimag(t *testing.T) {
	if imag(complex(1, 2)) != Bimag(complex(1, 2)) {
		t.Errorf("Bimag behavior should be equal to imag")
	}
	if imag(complex64(complex(1, 2))) != Bimag(complex64(complex(1, 2))) {
		t.Errorf("Bimag behavior should be equal to imag")
	}
}

func TestBlen(t *testing.T) {
	if len([]int{1, 2, 3}) != Blen([]int{1, 2, 3}) {
		t.Errorf("Blen behavior should be equal to len")
	}
}

func TestBnew(t *testing.T) {
	var expected interface{}
	var got interface{}

	// builtin types
	expected = new(string)
	got = Bnew(RVOf(new(string)).Elem().Type())
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("expected=%#v, got=%#v", expected, got)
	}

	expected = new(uint32)
	got = Bnew(RVOf(new(uint32)).Elem().Type())
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("expected=%#v, got=%#v", expected, got)
	}

	expected = new(error)
	got = Bnew(RVOf(new(error)).Elem().Type())
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("expected=%#v, got=%#v", expected, got)
	}

	// slice, map
	expected = new([]string)
	got = Bnew(RVOf(new([]string)).Elem().Type())
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("expected=%#v, got=%#v", expected, got)
	}

	expected = new(map[string]string)
	got = Bnew(RVOf(new(map[string]string)).Elem().Type())
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("expected=%#v, got=%#v", expected, got)
	}

	// chan
	expected = new(chan int)
	got = Bnew(RVOf(new(chan int)).Elem().Type())
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("expected=%#v, got=%#v", expected, got)
	}

	// other type
	type foo struct{}
	expected = new(foo)
	got = Bnew(RVOf(new(foo)).Elem().Type())
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("expected=%#v, got=%#v", expected, got)
	}
}

func TestBreal(t *testing.T) {
	if real(complex(1, 2)) != Breal(complex(1, 2)) {
		t.Errorf("Breal behavior should be equal to real")
	}
	if real(complex64(complex(1, 2))) != Breal(complex64(complex(1, 2))) {
		t.Errorf("Breal behavior should be equal to real")
	}
}

func TestBmake(t *testing.T) {
	var expected interface{}
	var got interface{}

	// slice
	expected = make([]string, 12, 14)
	got = Bmake(RTOf([]string{}), 12, 14)
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("expected=%#v, got=%#v", expected, got)
	}

	expected = make([]string, 12)
	got = Bmake(RTOf([]string{}), 12, 12)
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("expected=%#v, got=%#v", expected, got)
	}
	if len(expected.([]string)) != len(got.([]string)) {
		t.Errorf("expected=%#v, got=%#v", len(expected.([]string)), len(got.([]string)))
	}
	if cap(expected.([]string)) != cap(got.([]string)) {
		t.Errorf("expected=%#v, got=%#v", cap(expected.([]string)), cap(got.([]string)))
	}

	// map
	expected = make(map[string]string)
	got = Bmake(RTOf(map[string]string{}))
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("expected=%#v, got=%#v", expected, got)
	}

	// chan
	var c interface{}
	c = Bmake(RTOf(make(chan int)))
	expected = 0
	got = cap(c.(chan int))
	if got != expected {
		t.Errorf("len(Bake(chan int)) must be %d, but got %d", expected, got)
	}
	c = Bmake(RTOf(make(chan int)), 1)
	expected = 1
	got = cap(c.(chan int))
	if got != expected {
		t.Errorf("len(Bake(chan int)) must be %d, but got %d", expected, got)
	}
}
