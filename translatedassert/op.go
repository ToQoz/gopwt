package translatedassert

// OpADD has nodoc
func OpADD(x interface{}, y interface{}) interface{} {
	switch x.(type) {
	case uint8:
		return x.(uint8) + y.(uint8)
	case uint16:
		return x.(uint16) + y.(uint16)
	case uint32:
		return x.(uint32) + y.(uint32)
	case uint64:
		return x.(uint64) + y.(uint64)
	case uint:
		return x.(uint) + y.(uint)
	case int8:
		return x.(int8) + y.(int8)
	case int16:
		return x.(int16) + y.(int16)
	case int32:
		return x.(int32) + y.(int32)
	case int64:
		return x.(int64) + y.(int64)
	case int:
		return x.(int) + y.(int)
	case float32:
		return x.(float32) + y.(float32)
	case float64:
		return x.(float64) + y.(float64)
	case complex64:
		return x.(complex64) + y.(complex64)
	case complex128:
		return x.(complex128) + y.(complex128)
	case string:
		return x.(string) + y.(string)
	}

	panic("+ can take integers, floats, complex values, strings")
}

// OpSUB has nodoc
func OpSUB(x interface{}, y interface{}) interface{} {
	switch x.(type) {
	case uint8:
		return x.(uint8) - y.(uint8)
	case uint16:
		return x.(uint16) - y.(uint16)
	case uint32:
		return x.(uint32) - y.(uint32)
	case uint64:
		return x.(uint64) - y.(uint64)
	case uint:
		return x.(uint) - y.(uint)
	case int8:
		return x.(int8) - y.(int8)
	case int16:
		return x.(int16) - y.(int16)
	case int32:
		return x.(int32) - y.(int32)
	case int64:
		return x.(int64) - y.(int64)
	case int:
		return x.(int) - y.(int)
	case float32:
		return x.(float32) - y.(float32)
	case float64:
		return x.(float64) - y.(float64)
	case complex64:
		return x.(complex64) - y.(complex64)
	case complex128:
		return x.(complex128) - y.(complex128)
	}

	panic("+ can take integers, floats, complex values")
}

// OpMUL has nodoc
func OpMUL(x interface{}, y interface{}) interface{} {
	switch x.(type) {
	case uint8:
		return x.(uint8) * y.(uint8)
	case uint16:
		return x.(uint16) * y.(uint16)
	case uint32:
		return x.(uint32) * y.(uint32)
	case uint64:
		return x.(uint64) * y.(uint64)
	case uint:
		return x.(uint) * y.(uint)
	case int8:
		return x.(int8) * y.(int8)
	case int16:
		return x.(int16) * y.(int16)
	case int32:
		return x.(int32) * y.(int32)
	case int64:
		return x.(int64) * y.(int64)
	case int:
		return x.(int) * y.(int)
	case float32:
		return x.(float32) * y.(float32)
	case float64:
		return x.(float64) * y.(float64)
	case complex64:
		return x.(complex64) * y.(complex64)
	case complex128:
		return x.(complex128) * y.(complex128)
	}

	panic("* can take integers, floats, complex values")
}

// OpQUO has nodoc
func OpQUO(x interface{}, y interface{}) interface{} {
	switch x.(type) {
	case uint8:
		return x.(uint8) / y.(uint8)
	case uint16:
		return x.(uint16) / y.(uint16)
	case uint32:
		return x.(uint32) / y.(uint32)
	case uint64:
		return x.(uint64) / y.(uint64)
	case uint:
		return x.(uint) / y.(uint)
	case int8:
		return x.(int8) / y.(int8)
	case int16:
		return x.(int16) / y.(int16)
	case int32:
		return x.(int32) / y.(int32)
	case int64:
		return x.(int64) / y.(int64)
	case int:
		return x.(int) / y.(int)
	case float32:
		return x.(float32) / y.(float32)
	case float64:
		return x.(float64) / y.(float64)
	case complex64:
		return x.(complex64) / y.(complex64)
	case complex128:
		return x.(complex128) / y.(complex128)
	}

	panic("/ can take integers, floats, complex values")
}

// OpREM has nodoc
func OpREM(x interface{}, y interface{}) interface{} {
	switch x.(type) {
	case uint8:
		return x.(uint8) % y.(uint8)
	case uint16:
		return x.(uint16) % y.(uint16)
	case uint32:
		return x.(uint32) % y.(uint32)
	case uint64:
		return x.(uint64) % y.(uint64)
	case uint:
		return x.(uint) % y.(uint)
	case int8:
		return x.(int8) % y.(int8)
	case int16:
		return x.(int16) % y.(int16)
	case int32:
		return x.(int32) % y.(int32)
	case int64:
		return x.(int64) % y.(int64)
	case int:
		return x.(int) % y.(int)
	}

	errm := "% can take integers"
	panic(errm)
}

// OpAND has nodoc
func OpAND(x interface{}, y interface{}) interface{} {
	switch x.(type) {
	case uint8:
		return x.(uint8) & y.(uint8)
	case uint16:
		return x.(uint16) & y.(uint16)
	case uint32:
		return x.(uint32) & y.(uint32)
	case uint64:
		return x.(uint64) & y.(uint64)
	case uint:
		return x.(uint) & y.(uint)
	case int8:
		return x.(int8) & y.(int8)
	case int16:
		return x.(int16) & y.(int16)
	case int32:
		return x.(int32) & y.(int32)
	case int64:
		return x.(int64) & y.(int64)
	case int:
		return x.(int) & y.(int)
	}

	panic("& can take integers")
}

// OpOR has nodoc
func OpOR(x interface{}, y interface{}) interface{} {
	switch x.(type) {
	case uint8:
		return x.(uint8) | y.(uint8)
	case uint16:
		return x.(uint16) | y.(uint16)
	case uint32:
		return x.(uint32) | y.(uint32)
	case uint64:
		return x.(uint64) | y.(uint64)
	case uint:
		return x.(uint) | y.(uint)
	case int8:
		return x.(int8) | y.(int8)
	case int16:
		return x.(int16) | y.(int16)
	case int32:
		return x.(int32) | y.(int32)
	case int64:
		return x.(int64) | y.(int64)
	case int:
		return x.(int) | y.(int)
	}

	panic("| can take integers")
}

// OpXOR has nodoc
func OpXOR(x interface{}, y interface{}) interface{} {
	switch x.(type) {
	case uint8:
		return x.(uint8) ^ y.(uint8)
	case uint16:
		return x.(uint16) ^ y.(uint16)
	case uint32:
		return x.(uint32) ^ y.(uint32)
	case uint64:
		return x.(uint64) ^ y.(uint64)
	case uint:
		return x.(uint) ^ y.(uint)
	case int8:
		return x.(int8) ^ y.(int8)
	case int16:
		return x.(int16) ^ y.(int16)
	case int32:
		return x.(int32) ^ y.(int32)
	case int64:
		return x.(int64) ^ y.(int64)
	case int:
		return x.(int) ^ y.(int)
	}

	panic("^ can take integers")
}

// OpANDNOT has nodoc
func OpANDNOT(x interface{}, y interface{}) interface{} { // nolint
	switch x.(type) {
	case uint8:
		return x.(uint8) &^ y.(uint8)
	case uint16:
		return x.(uint16) &^ y.(uint16)
	case uint32:
		return x.(uint32) &^ y.(uint32)
	case uint64:
		return x.(uint64) &^ y.(uint64)
	case uint:
		return x.(uint) &^ y.(uint)
	case int8:
		return x.(int8) &^ y.(int8)
	case int16:
		return x.(int16) &^ y.(int16)
	case int32:
		return x.(int32) &^ y.(int32)
	case int64:
		return x.(int64) &^ y.(int64)
	case int:
		return x.(int) &^ y.(int)
	}

	panic("&^ can take integers")
}

// OpSHL has nodoc
func OpSHL(x interface{}, y interface{}) interface{} {
	switch x.(type) {
	case uint8:
		x := x.(uint8)

		switch y.(type) {
		case uint8:
			return x << y.(uint8)
		case uint16:
			return x << y.(uint16)
		case uint32:
			return x << y.(uint32)
		case uint64:
			return x << y.(uint64)
		case uint:
			return x << y.(uint)
		}
	case uint16:
		x := x.(uint16)

		switch y.(type) {
		case uint8:
			return x << y.(uint8)
		case uint16:
			return x << y.(uint16)
		case uint32:
			return x << y.(uint32)
		case uint64:
			return x << y.(uint64)
		case uint:
			return x << y.(uint)
		}
	case uint32:
		x := x.(uint32)

		switch y.(type) {
		case uint8:
			return x << y.(uint8)
		case uint16:
			return x << y.(uint16)
		case uint32:
			return x << y.(uint32)
		case uint64:
			return x << y.(uint64)
		case uint:
			return x << y.(uint)
		}
	case uint64:
		x := x.(uint64)

		switch y.(type) {
		case uint8:
			return x << y.(uint8)
		case uint16:
			return x << y.(uint16)
		case uint32:
			return x << y.(uint32)
		case uint64:
			return x << y.(uint64)
		case uint:
			return x << y.(uint)
		}
	case uint:
		x := x.(uint)

		switch y.(type) {
		case uint8:
			return x << y.(uint8)
		case uint16:
			return x << y.(uint16)
		case uint32:
			return x << y.(uint32)
		case uint64:
			return x << y.(uint64)
		case uint:
			return x << y.(uint)
		}
	case int8:
		x := x.(int8)

		switch y.(type) {
		case uint8:
			return x << y.(uint8)
		case uint16:
			return x << y.(uint16)
		case uint32:
			return x << y.(uint32)
		case uint64:
			return x << y.(uint64)
		case uint:
			return x << y.(uint)
		}
	case int16:
		x := x.(int16)

		switch y.(type) {
		case uint8:
			return x << y.(uint8)
		case uint16:
			return x << y.(uint16)
		case uint32:
			return x << y.(uint32)
		case uint64:
			return x << y.(uint64)
		case uint:
			return x << y.(uint)
		}
	case int32:
		x := x.(int32)

		switch y.(type) {
		case uint8:
			return x << y.(uint8)
		case uint16:
			return x << y.(uint16)
		case uint32:
			return x << y.(uint32)
		case uint64:
			return x << y.(uint64)
		case uint:
			return x << y.(uint)
		}
	case int64:
		x := x.(int64)

		switch y.(type) {
		case uint8:
			return x << y.(uint8)
		case uint16:
			return x << y.(uint16)
		case uint32:
			return x << y.(uint32)
		case uint64:
			return x << y.(uint64)
		case uint:
			return x << y.(uint)
		}
	case int:
		x := x.(int)

		switch y.(type) {
		case uint8:
			return x << y.(uint8)
		case uint16:
			return x << y.(uint16)
		case uint32:
			return x << y.(uint32)
		case uint64:
			return x << y.(uint64)
		case uint:
			return x << y.(uint)
		}
	}

	panic("<< can take (left)integer, (right)unsigned integer")
}

// OpSHR has nodoc
func OpSHR(x interface{}, y interface{}) interface{} {
	switch x.(type) {
	case uint8:
		x := x.(uint8)

		switch y.(type) {
		case uint8:
			return x >> y.(uint8)
		case uint16:
			return x >> y.(uint16)
		case uint32:
			return x >> y.(uint32)
		case uint64:
			return x >> y.(uint64)
		case uint:
			return x >> y.(uint)
		}
	case uint16:
		x := x.(uint16)

		switch y.(type) {
		case uint8:
			return x >> y.(uint8)
		case uint16:
			return x >> y.(uint16)
		case uint32:
			return x >> y.(uint32)
		case uint64:
			return x >> y.(uint64)
		case uint:
			return x >> y.(uint)
		}
	case uint32:
		x := x.(uint32)

		switch y.(type) {
		case uint8:
			return x >> y.(uint8)
		case uint16:
			return x >> y.(uint16)
		case uint32:
			return x >> y.(uint32)
		case uint64:
			return x >> y.(uint64)
		case uint:
			return x >> y.(uint)
		}
	case uint64:
		x := x.(uint64)

		switch y.(type) {
		case uint8:
			return x >> y.(uint8)
		case uint16:
			return x >> y.(uint16)
		case uint32:
			return x >> y.(uint32)
		case uint64:
			return x >> y.(uint64)
		case uint:
			return x >> y.(uint)
		}
	case uint:
		x := x.(uint)

		switch y.(type) {
		case uint8:
			return x >> y.(uint8)
		case uint16:
			return x >> y.(uint16)
		case uint32:
			return x >> y.(uint32)
		case uint64:
			return x >> y.(uint64)
		case uint:
			return x >> y.(uint)
		}
	case int8:
		x := x.(int8)

		switch y.(type) {
		case uint8:
			return x >> y.(uint8)
		case uint16:
			return x >> y.(uint16)
		case uint32:
			return x >> y.(uint32)
		case uint64:
			return x >> y.(uint64)
		case uint:
			return x >> y.(uint)
		}
	case int16:
		x := x.(int16)

		switch y.(type) {
		case uint8:
			return x >> y.(uint8)
		case uint16:
			return x >> y.(uint16)
		case uint32:
			return x >> y.(uint32)
		case uint64:
			return x >> y.(uint64)
		case uint:
			return x >> y.(uint)
		}
	case int32:
		x := x.(int32)

		switch y.(type) {
		case uint8:
			return x >> y.(uint8)
		case uint16:
			return x >> y.(uint16)
		case uint32:
			return x >> y.(uint32)
		case uint64:
			return x >> y.(uint64)
		case uint:
			return x >> y.(uint)
		}
	case int64:
		x := x.(int64)

		switch y.(type) {
		case uint8:
			return x >> y.(uint8)
		case uint16:
			return x >> y.(uint16)
		case uint32:
			return x >> y.(uint32)
		case uint64:
			return x >> y.(uint64)
		case uint:
			return x >> y.(uint)
		}
	case int:
		x := x.(int)

		switch y.(type) {
		case uint8:
			return x >> y.(uint8)
		case uint16:
			return x >> y.(uint16)
		case uint32:
			return x >> y.(uint32)
		case uint64:
			return x >> y.(uint64)
		case uint:
			return x >> y.(uint)
		}
	}

	panic(">> can take (left)integer, (right)unsigned integer")
}

// OpLAND has nodoc
func OpLAND(x interface{}, y interface{}) bool {
	switch x.(type) {
	case bool:
		return x.(bool) && y.(bool)
	}

	panic("&& can take bool")
}

// OpLOR has nodoc
func OpLOR(x interface{}, y interface{}) bool {
	switch x.(type) {
	case bool:
		return x.(bool) || y.(bool)
	}

	panic("|| can bool")
}
