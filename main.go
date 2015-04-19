package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

var (
	termw   = 0
	verbose = flag.Bool("v", false, "This will be passed to `go test`")
)

func main() {
	if err := doMain(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
		return
	}
}
func doMain() error {
	flag.Parse()

	termw = getTermCols(os.Stdin.Fd())

	tempGoPath, err := ioutil.TempDir(os.TempDir(), "")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempGoPath)

	root := flag.Arg(0)

	pkgInfo, err := newPackageInfo(root)
	if err != nil {
		return err
	}

	err = rewrite(tempGoPath, pkgInfo)
	if err != nil {
		return err
	}

	err = runTest(tempGoPath, pkgInfo, os.Stdout, os.Stderr)
	if err != nil {
		return err
	}

	return nil
}

func rewrite(tempGoPath string, pkgInfo *packageInfo) error {
	tempGoSrcDir := filepath.Join(tempGoPath, "src")

	err := filepath.Walk(pkgInfo.dirPath, func(path string, fInfo os.FileInfo, err error) error {
		if fInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
			return nil
		}

		if !fInfo.IsDir() {
			return nil
		}

		c, err := containsGoFile(path)
		if err != nil {
			return err
		}
		if !c {
			return filepath.SkipDir
		}

		rel, err := filepath.Rel(pkgInfo.dirPath, path)
		if err != nil {
			return err
		}

		if strings.Split(rel, "/")[0] == "testdata" {
			return filepath.SkipDir
		}

		if rel != "." {
			if filepath.HasPrefix(rel, ".") {
				return filepath.SkipDir
			}

			if !pkgInfo.recursive {
				return filepath.SkipDir
			}
		}

		importPath := filepath.Join(pkgInfo.importPath, rel)

		err = os.MkdirAll(filepath.Join(tempGoSrcDir, importPath), os.ModePerm)
		if err != nil {
			return err
		}

		err = rewritePackage(path, importPath, tempGoSrcDir)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func runTest(goPath string, pkgInfo *packageInfo, stdout, stderr io.Writer) error {
	err := os.Setenv("GOPATH", goPath+":"+os.Getenv("GOPATH"))
	if err != nil {
		return err
	}

	cmd := exec.Command("go", "test")
	if *verbose {
		cmd.Args = append(cmd.Args, "-v")
	}
	cmd.Args = append(cmd.Args, pkgInfo.ToGoTestArg())
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

func containsGoFile(dir string) (bool, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return false, err
	}

	for _, f := range files {
		if isGoFile(f) {
			return true, nil
		}
	}

	return false, nil
}

func isGoFile(f os.FileInfo) bool {
	name := f.Name()
	return !f.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go")
}

func getTermCols(fd uintptr) int {
	var sz = struct {
		_    uint16
		cols uint16
		_    uint16
		_    uint16
	}{}
	_, _, _ = syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&sz)))
	return int(sz.cols)
}
