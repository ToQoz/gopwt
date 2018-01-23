// +build !windows

package gopwt

import (
	"os"
	"syscall"
	"unsafe"

	"github.com/ToQoz/gopwt/translator"
)

func setTermCols() {
	translator.TermWidth(getTermCols(os.Stdout.Fd()))
}

func getTermCols(fd uintptr) int {
	var sz = struct {
		_    uint16
		cols uint16
		_    uint16
		_    uint16
	}{}
	_, _, _ = syscall.Syscall(
		syscall.SYS_IOCTL,
		fd,
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(&sz)),
	)
	return int(sz.cols)
}
