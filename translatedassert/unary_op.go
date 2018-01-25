package translatedassert

import (
	"reflect"
)

// UnaryOpADD has nodoc
func UnaryOpADD(x interface{}) interface{} {
	switch x.(type) {
	case uint8:
		return +x.(uint8)
	case uint16:
		return +x.(uint16)
	case uint32:
		return +x.(uint32)
	case uint64:
		return +x.(uint64)
	case uint:
		return +x.(uint)
	case int8:
		return +x.(int8)
	case int16:
		return +x.(int16)
	case int32:
		return +x.(int32)
	case int64:
		return +x.(int64)
	case int:
		return +x.(int)
	case float32:
		return +x.(float32)
	case float64:
		return +x.(float64)
	case complex64:
		return +x.(complex64)
	case complex128:
		return +x.(complex128)
	}
	panic("unary(+) can take integers, floats, complex values")
}

// UnaryOpSUB has nodoc
func UnaryOpSUB(x interface{}) interface{} {
	switch x.(type) {
	case uint8:
		return -x.(uint8)
	case uint16:
		return -x.(uint16)
	case uint32:
		return -x.(uint32)
	case uint64:
		return -x.(uint64)
	case uint:
		return -x.(uint)
	case int8:
		return -x.(int8)
	case int16:
		return -x.(int16)
	case int32:
		return -x.(int32)
	case int64:
		return -x.(int64)
	case int:
		return -x.(int)
	case float32:
		return -x.(float32)
	case float64:
		return -x.(float64)
	case complex64:
		return -x.(complex64)
	case complex128:
		return -x.(complex128)
	}
	panic("unary(-) can take integers, floats, complex values")
}

// UnaryOpNOT has nodoc
func UnaryOpNOT(x interface{}) interface{} {
	switch x.(type) {
	case bool:
		return !x.(bool)
	}
	panic("unary(!) can take bool")
}

// UnaryOpXOR has nodoc
func UnaryOpXOR(x interface{}) interface{} {
	switch x.(type) {
	case uint8:
		return ^x.(uint8)
	case uint16:
		return ^x.(uint16)
	case uint32:
		return ^x.(uint32)
	case uint64:
		return ^x.(uint64)
	case uint:
		return ^x.(uint)
	case int8:
		return ^x.(int8)
	case int16:
		return ^x.(int16)
	case int32:
		return ^x.(int32)
	case int64:
		return ^x.(int64)
	case int:
		return ^x.(int)
	}
	panic("unary(^) can take integers")
}

// UnaryOpARROW has nodoc
func UnaryOpARROW(x interface{}) interface{} {
	r, _ := reflect.ValueOf(x).Recv()
	return r.Interface()
}
