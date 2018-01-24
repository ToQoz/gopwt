package gopwt

import (
	"flag"
	"fmt"
	"go/build"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"syscall"

	"github.com/ToQoz/gopwt/translator"
	"github.com/mattn/go-isatty"
)

var (
	verbose  = false
	testdata = flag.String("testdata", "testdata", "name of test data directories. e.g. -testdata testdata,migrations")
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

	translator.Verbose(verbose)
	translator.Testdata(*testdata)
	if isatty.IsTerminal(os.Stdout.Fd()) {
		setTermCols()
	}
	if wd, err := os.Getwd(); err == nil {
		translator.WorkingDir(wd + string(filepath.Separator))
	}

	tmpGopath, importpath, err := translator.Translate(flag.Arg(0))
	if os.Getenv("GOPWT_DEBUG") == "" { // NOTE: don't remove translated files on GOPWT_DEBUG=1
		defer os.RemoveAll(tmpGopath)
	}
	if err != nil {
		return err
	}

	return runTest(tmpGopath, importpath, os.Stdout, os.Stderr)
}

func runTest(gopath string, importpath string, stdout, stderr io.Writer) error {
	if os.Getenv("GOPATH") == "" {
		// NOTE:
		// Without this line, we got error() to run. (os=windows, version=1.9.3, repo=ToQoz/gopwt, command=go test./...)
		err := os.Setenv("GOPATH", gopath+string(filepath.ListSeparator)+build.Default.GOPATH)
		if err != nil {
			return err
		}
	} else {
		err := os.Setenv("GOPATH", gopath+string(filepath.ListSeparator)+os.Getenv("GOPATH"))
		if err != nil {
			return err
		}
	}
	cmd := exec.Command("go", "test")

	if verbose {
		cmd.Args = append(cmd.Args, "-v")
	}
	cmd.Dir = path.Join(gopath, "src", importpath)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}
