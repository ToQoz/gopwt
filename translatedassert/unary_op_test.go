package translatedassert

import "testing"

type UnaryExample struct {
	x        interface{}
	expected interface{}
}

func TestUnaryOpADD(t *testing.T) {
	tests := []UnaryExample{
		{1, 1},
		{uint(1), uint(1)},
		{uint8(1), uint8(1)},
		{uint16(1), uint16(1)},
		{uint32(1), uint32(1)},
		{uint64(1), uint64(1)},
		{int8(1), int8(1)},
		{int16(1), int16(1)},
		{int32(1), int32(1)},
		{int64(1), int64(1)},
		{float32(1), float32(1)},
		{float64(1), float64(1)},
		{complex(1, 0), complex(1, 0)},
		{complex64(complex(1, 0)), complex64(complex(1, 0))},
	}

	for _, test := range tests {
		got := UnaryOpADD(test.x)
		if !eq(test.expected, got) {
			t.Errorf(`UnaryOpADD(%v) = %v, but got %v`, test.x, test.expected, got)
		}
	}
}

func TestUnaryOpSUB(t *testing.T) {
	var u uint = 1
	var u8 uint8 = 1
	var u16 uint16 = 1
	var u32 uint32 = 1
	var u64 uint64 = 1
	tests := []UnaryExample{
		{1, -1},
		{int8(1), -int8(1)},
		{int16(1), -int16(1)},
		{int32(1), -int32(1)},
		{int64(1), -int64(1)},
		{u, -u},
		{u8, -u8},
		{u16, -u16},
		{u32, -u32},
		{u64, -u64},
		{float32(1), -float32(1)},
		{float64(1), -float64(1)},
		{complex(1, 0), -complex(1, 0)},
		{complex64(complex(1, 0)), -complex64(complex(1, 0))},
	}

	for _, test := range tests {
		got := UnaryOpSUB(test.x)
		if !eq(test.expected, got) {
			t.Errorf(`UnaryOpSUB(%v) = %v, but got %v`, test.x, test.expected, got)
		}
	}
}

func TestUnaryOpNOT(t *testing.T) {
	tests := []UnaryExample{
		{true, false},
		{false, true},
	}

	for _, test := range tests {
		got := UnaryOpNOT(test.x)
		if !eq(test.expected, got) {
			t.Errorf(`UnaryOpNOT(%v) = %v, but got %v`, test.x, test.expected, got)
		}
	}
}

func TestUnaryOpXOR(t *testing.T) {
	tests := []UnaryExample{
		{-1, ^-1},
		{0, ^0},
		{1, ^1},
		{uint(1), ^uint(1)},
		{uint8(1), ^uint8(1)},
		{uint16(1), ^uint16(1)},
		{uint32(1), ^uint32(1)},
		{uint64(1), ^uint64(1)},
		{int8(1), ^int8(1)},
		{int16(1), ^int16(1)},
		{int32(1), ^int32(1)},
		{int64(1), ^int64(1)},
	}

	for _, test := range tests {
		got := UnaryOpXOR(test.x)
		if !eq(test.expected, got) {
			t.Errorf(`UnaryOpXOR(%v) = %v, but got %v`, test.x, test.expected, got)
		}
	}
}
