package translatedassert

func isEitherComplex64(x, y interface{}) bool {
	_, xok := x.(complex64)
	_, yok := y.(complex64)
	return xok || yok
}

func isEitherComplex128(x, y interface{}) bool {
	_, xok := x.(complex128)
	_, yok := y.(complex128)
	return xok || yok
}

func isEitherFloat32(x, y interface{}) bool {
	_, xok := x.(float32)
	_, yok := y.(float32)
	return xok || yok
}

func isEitherFloat64(x, y interface{}) bool {
	_, xok := x.(float64)
	_, yok := y.(float64)
	return xok || yok
}

func isEitherUint(x, y interface{}) bool {
	_, xok := x.(uint)
	_, yok := y.(uint)
	return xok || yok
}

func isEitherUint8(x, y interface{}) bool {
	_, xok := x.(uint8)
	_, yok := y.(uint8)
	return xok || yok
}

func isEitherUint16(x, y interface{}) bool {
	_, xok := x.(uint16)
	_, yok := y.(uint16)
	return xok || yok
}

func isEitherUint32(x, y interface{}) bool {
	_, xok := x.(uint32)
	_, yok := y.(uint32)
	return xok || yok
}

func isEitherUint64(x, y interface{}) bool {
	_, xok := x.(uint64)
	_, yok := y.(uint64)
	return xok || yok
}

func asUint(x interface{}) uint {
	switch x.(type) {
	case uint:
		return x.(uint)
	case int:
		return uint(x.(int))
	}
	panic("fail to uint")
}

func asUint8(x interface{}) uint8 {
	switch x.(type) {
	case uint8:
		return x.(uint8)
	case int:
		return uint8(x.(int))
	}
	panic("fail to uint8")
}

func asUint16(x interface{}) uint16 {
	switch x.(type) {
	case uint16:
		return x.(uint16)
	case int:
		return uint16(x.(int))
	}
	panic("fail to uint16")
}

func asUint32(x interface{}) uint32 {
	switch x.(type) {
	case uint32:
		return x.(uint32)
	case int:
		return uint32(x.(int))
	}
	panic("fail to uint32")
}

func asUint64(x interface{}) uint64 {
	switch x.(type) {
	case uint64:
		return x.(uint64)
	case int:
		return uint64(x.(int))
	}
	panic("fail to uint64")
}

func asComplex64(x interface{}) complex64 {
	switch x.(type) {
	case complex64:
		return x.(complex64)
	case int32:
		return complex(float32(x.(int32)), 0)
	case int64:
		return complex(float32(x.(int64)), 0)
	case int:
		return complex(float32(x.(int)), 0)
	case float32:
		return complex(x.(float32), 0)
	}
	panic("fail to asComplex64")
}

func asComplex128(x interface{}) complex128 {
	switch x.(type) {
	case complex128:
		return x.(complex128)
	case int32:
		return complex(float64(x.(int32)), 0)
	case int64:
		return complex(float64(x.(int64)), 0)
	case int:
		return complex(float64(x.(int)), 0)
	case float64:
		return complex(x.(float64), 0)
	}
	panic("fail to asComplex128")
}

func asFloat32(x interface{}) float32 {
	switch x.(type) {
	case int32:
		return float32(x.(int32))
	case int64:
		return float32(x.(int64))
	case int:
		return float32(x.(int))
	case float32:
		return x.(float32)
	case float64:
		return float32(x.(float64))
	}
	panic("fail to float32")
}

func asFloat64(x interface{}) float64 {
	switch x.(type) {
	case int32:
		return float64(x.(int32))
	case int64:
		return float64(x.(int64))
	case int:
		return float64(x.(int))
	case float32:
		return float64(x.(float32))
	case float64:
		return x.(float64)
	}
	panic("fail to float64")
}
