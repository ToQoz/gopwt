package translatedassert

import (
	"fmt"
	"reflect"
)

// --- builtin funcs ---
// func append(slice []Type, elems ...Type) []Type
// func cap(v Type) int
// [NP]func close(c chan<- Type)
// func complex(r, i FloatType) ComplexType
// func copy(dst, src []Type) int
// [NP]func delete(m map[Type]Type1, key Type)
// func imag(c ComplexType) FloatType
// func len(v Type) int
// func make(Type, size IntegerType) Type
// func new(Type) *Type
// [NP]func panic(v interface{})
// [NP]func print(args ...Type)
// [NP]func println(args ...Type)
// func real(c ComplexType) FloatType
// [NP]func recover() interface{}

// Bappend has nodoc
func Bappend(s interface{}, x ...interface{}) interface{} {
	args := []reflect.Value{}
	for _, e := range x {
		args = append(args, reflect.ValueOf(e))
	}

	return reflect.Append(reflect.ValueOf(s), args...).Interface()
}

// Bcap has nodoc
func Bcap(s interface{}) int {
	return reflect.ValueOf(s).Cap()
}

// Bcomplex has nodoc
func Bcomplex(r, i interface{}) (c interface{}) {
	switch r.(type) {
	case float32:
		c = complex(r.(float32), i.(float32))
		return
	case float64:
		c = complex(r.(float64), i.(float64))
		return
	}

	panic("complex can take floats")
}

// Bcopy has nodoc
func Bcopy(dst, src interface{}) int {
	return reflect.Copy(reflect.ValueOf(dst), reflect.ValueOf(src))
}

// Bimag has nodoc
func Bimag(c interface{}) interface{} {
	switch c.(type) {
	case complex64:
		return imag(c.(complex64))
	case complex128:
		return imag(c.(complex128))
	}

	panic("imag can take ComplexType")
}

// Blen has nodoc
func Blen(v interface{}) int {
	return reflect.ValueOf(v).Len()
}

// Bmake has nodoc
// convert reflect's func
// func MakeChan(typ Type, buffer int) Value
// func MakeSlice(typ Type, len, cap int) Value
// func MakeMap(typ Type) Value
func Bmake(typ reflect.Type, args ...interface{}) interface{} {
	usage := `Usage of make
Call             Type T     Result

make(T, n)       slice      slice of type T with length n and capacity n
make(T, n, m)    slice      slice of type T with length n and capacity m

make(T)          map        map of type T
make(T, n)       map        map of type T with initial space for n elements

make(T)          channel    unbuffered channel of type T
make(T, n)       channel    buffered channel of type T, buffer size n
`

	switch typ.Kind() {
	case reflect.Slice, reflect.Array:
		var _len, _cap int

		if len(args) < 1 {
			panic(fmt.Errorf("missing len argument to make\n%s", usage))
		}

		if i, ok := args[0].(int); ok {
			_len = i
		} else {
			panic(fmt.Errorf("cannot convert %#v to type int\n%s", args[1], usage))
		}

		if len(args) >= 2 {
			if i, ok := args[1].(int); ok {
				_cap = i
			} else {
				panic(fmt.Errorf("cannot convert %#v to type int\n%s", args[2], usage))
			}
		} else {
			_cap = _len
		}

		return reflect.MakeSlice(typ, _len, _cap).Interface()
	case reflect.Map:
		return reflect.MakeMap(typ).Interface()
	case reflect.Chan:
		var _buf int

		if len(args) >= 1 {
			if i, ok := args[0].(int); ok {
				_buf = i
			} else {
				panic(fmt.Errorf("cannot convert %#v to type int\n%s", args[0], usage))
			}
		} else {
			_buf = 0
		}

		return reflect.MakeChan(typ, _buf).Interface()
	}

	panic(fmt.Errorf("cannot make %#v\n%s", args[0], usage))
}

// Bnew has nodoc
func Bnew(a reflect.Type) interface{} {
	return reflect.New(a).Interface()
}

// Breal has nodoc
func Breal(c interface{}) interface{} {
	switch c.(type) {
	case complex64:
		return real(c.(complex64))
	case complex128:
		return real(c.(complex128))
	}

	panic("real can take ComplexType")
}
