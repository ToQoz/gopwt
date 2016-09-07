package gopwt

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/mattn/go-isatty"
)

var (
	termw    = 0
	testdata = flag.String("testdata", "testdata", "name of test data directories. e.g. -testdata testdata,migrations")
	verbose  = false
)

func Empower() {
	if os.Getenv("GOPWT_OFF") != "" {
		return
	}

	if err := doMain(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())

		if exiterr, ok := err.(*exec.ExitError); ok {
			if s, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				os.Exit(s.ExitStatus())
			} else {
				panic(fmt.Errorf("Unimplemented for system where exec.ExitError.Sys() is not syscall.WaitStatus."))
			}
		}

		os.Exit(127)
	}
	os.Exit(0)
}
func doMain() error {
	if runtime.Version() == "go1.4" {
		return fmt.Errorf("go1.4 is not supported. please bump to go1.4.1 or later")
	}

	if !flag.Parsed() {
		flag.Parse()
	}

	flag.VisitAll(func(f *flag.Flag) {
		if f.Name == "test.v" {
			if f.Value.String() != "false" {
				verbose = true
			}
		}
	})

	if isatty.IsTerminal(os.Stdout.Fd()) {
		termw = getTermCols(os.Stdin.Fd())
	}

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

		files, err := ioutil.ReadDir(path)
		if err != nil {
			return err
		}
		if !containsGoFile(files) {
			// sub-packages maybe have gofiles, even if itself don't has gofiles
			if containsDirectory(files) {
				return nil
			}
			return filepath.SkipDir
		}

		rel, err := filepath.Rel(pkgInfo.dirPath, path)
		if err != nil {
			return err
		}

		for _, tdata := range strings.Split(*testdata, ",") {
			if strings.Split(rel, "/")[0] == tdata {
				return filepath.SkipDir
			}
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

	if verbose {
		cmd.Args = append(cmd.Args, "-v")
	}
	cmd.Dir = path.Join(goPath, "src", pkgInfo.importPath)
	// cmd.Args = append(cmd.Args, pkgInfo.ToGoTestArg())
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}
