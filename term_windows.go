// +build windows

package gopwt

import (
	"syscall"
	"unsafe"

	"github.com/ToQoz/gopwt/translator"
)

var getConsoleScreenBufferInfo = syscall.NewLazyDLL("kernel32.dll").NewProc("GetConsoleScreenBufferInfo")

func setTermCols() {
	conout, err := syscall.Open("CONOUT$", syscall.O_RDWR, 0)
	if err == nil {
		translator.TermWidth(getTermCols(uintptr(conout)))
	}
}

func getTermCols(fd uintptr) int {
	var sz = struct {
		size struct {
			x int16
			y int16
		}
		_ struct {
			_ uint16
			_ uint16
		}
		_ uint16
		_ struct {
			_ int16
			_ int16
			_ int16
			_ int16
		}
		_ struct {
			x int16
			y int16
		}
	}{}
	_, _, _ = syscall.Syscall(
		getConsoleScreenBufferInfo.Addr(),
		2,
		fd,
		uintptr(unsafe.Pointer(&sz)),
		0,
	)
	return int(sz.size.x)
}
