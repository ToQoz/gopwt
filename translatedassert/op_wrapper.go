package translatedassert

import (
	"fmt"
	"reflect"
)

var binExprKindOrders = []reflect.Kind{
	reflect.String,
	reflect.Bool,
	reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
	reflect.Complex64, reflect.Complex128,
	reflect.Float32, reflect.Float64,
	reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
}

var binExprTypes = map[reflect.Kind]reflect.Type{
	reflect.Bool:    reflect.TypeOf(true),
	reflect.String:  reflect.TypeOf(string("")),
	reflect.Float32: reflect.TypeOf(float32(0)),
	reflect.Float64: reflect.TypeOf(float64(0)),
	reflect.Int:     reflect.TypeOf(int(0)),
	reflect.Int8:    reflect.TypeOf(int8(0)),
	reflect.Int16:   reflect.TypeOf(int16(0)),
	reflect.Int32:   reflect.TypeOf(int32(0)),
	reflect.Int64:   reflect.TypeOf(int64(0)),
	reflect.Uint:    reflect.TypeOf(uint(0)),
	reflect.Uint8:   reflect.TypeOf(uint8(0)),
	reflect.Uint16:  reflect.TypeOf(uint16(0)),
	reflect.Uint32:  reflect.TypeOf(uint32(0)),
	reflect.Uint64:  reflect.TypeOf(uint64(0)),
}

func Op(op string, _x, _y interface{}) interface{} {
	x := reflect.ValueOf(_x)
	xt := x.Kind()
	y := reflect.ValueOf(_y)
	yt := y.Kind()

	var kind reflect.Kind
	for _, k := range binExprKindOrders {
		if xt == k || yt == k {
			kind = k
			goto VALID_KIND
		}
	}
	panic(fmt.Sprintf("unexpected kind for %s: x=%v, y=%v", op, x, y))

VALID_KIND:
	if kind == reflect.Complex128 { // NOTE: handle 1 + 2i | 2i + 1
		if xt != yt {
			if xt == kind {
				y = reflect.ValueOf(newComplex(_y))
			} else {
				x = reflect.ValueOf(newComplex(_x))
			}
		}
	} else if kind == reflect.Complex64 { // NOTE: handle complex64(complex(1, 2)) + 1 | 1 + complex64(complex(1, 2))
		if xt != yt {
			if xt == kind {
				y = reflect.ValueOf(complex64(newComplex(_y)))
			} else {
				x = reflect.ValueOf(complex64(newComplex(_x)))
			}
		}
	} else if kind == reflect.Float32 || kind == reflect.Float64 { // NOTE: handle  1 + 2.3 | 2.3 + 1
		if xt != yt {
			if xt == kind {
				y = y.Convert(reflect.TypeOf(_x))
			} else {
				x = x.Convert(reflect.TypeOf(_y))
			}
		}
	}

	// NOTE: handle myint(1) + myint(2) | myint(1) + 2 | 2 + myint(1)
	fn := reflect.ValueOf(ops[fmt.Sprintf("Op%s_%s", op, kind)])
	isXUserDefined := x.Type().String() != kind.String()
	isYUserDefined := y.Type().String() != kind.String()
	if isXUserDefined && isYUserDefined {
		return fn.Call([]reflect.Value{x.Convert(binExprTypes[kind]), y.Convert(binExprTypes[kind])})[0].Convert(x.Type()).Interface()
	} else if isXUserDefined {
		return fn.Call([]reflect.Value{x.Convert(binExprTypes[kind]), y})[0].Convert(x.Type()).Interface()
	} else if isYUserDefined {
		return fn.Call([]reflect.Value{x, y.Convert(binExprTypes[kind])})[0].Convert(y.Type()).Interface()
	}
	return fn.Call([]reflect.Value{x, y})[0].Interface()
}

func OpShift(op string, _x, _y interface{}) interface{} {
	x := reflect.ValueOf(_x)
	var y reflect.Value
	switch _y.(type) {
	case int: // NOTE: handle 1 << 1
		y = reflect.ValueOf(uint(_y.(int)))
	default:
		y = reflect.ValueOf(_y)
	}

	fn := reflect.ValueOf(ops[fmt.Sprintf("Op%s_%s_%s", op, x.Kind(), y.Kind())])

	// NOTE: handle myint(1) << myint(2) | myint(1) << 1 |  1 << myuint(1)
	isXUserDefined := x.Type().String() != x.Kind().String()
	isYUserDefined := y.Type().String() != y.Kind().String()
	if isXUserDefined && isYUserDefined {
		return fn.Call([]reflect.Value{x.Convert(binExprTypes[x.Kind()]), y.Convert(binExprTypes[y.Kind()])})[0].Convert(x.Type()).Interface()
	} else if isXUserDefined {
		return fn.Call([]reflect.Value{x.Convert(binExprTypes[x.Kind()]), y})[0].Convert(x.Type()).Interface()
	} else if isYUserDefined {
		return fn.Call([]reflect.Value{x, y.Convert(binExprTypes[y.Kind()])})[0].Interface()
	}

	return fn.Call([]reflect.Value{x, y})[0].Interface()
}

func newComplex(a interface{}) complex128 {
	switch a.(type) {
	case float32:
		return complex(float64(a.(float32)), 0)
	case float64:
		return complex(a.(float64), 0)
	case int:
		return complex(float64(a.(int)), 0)
	case int8:
		return complex(float64(a.(int8)), 0)
	case int16:
		return complex(float64(a.(int16)), 0)
	case int32:
		return complex(float64(a.(int32)), 0)
	case int64:
		return complex(float64(a.(int64)), 0)
	}

	panic(fmt.Sprintf("failt to convert %s to complex", a))
}
