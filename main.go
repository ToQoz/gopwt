package gopwt

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"runtime"
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

	translator.Verbose(verbose)
	if isatty.IsTerminal(os.Stdout.Fd()) {
		translator.TermWidth(getTermCols(os.Stdin.Fd()))
	}
	translator.Testdata(*testdata)

	gopath, err := ioutil.TempDir(os.TempDir(), "")
	if err != nil {
		return err
	}
	defer os.RemoveAll(gopath)

	importpath, err := translator.Translate(gopath, flag.Arg(0))
	if err != nil {
		return err
	}

	err = runTest(gopath, importpath, os.Stdout, os.Stderr)
	if err != nil {
		return err
	}

	return nil
}

func runTest(goPath string, importPath string, stdout, stderr io.Writer) error {
	err := os.Setenv("GOPATH", goPath+":"+os.Getenv("GOPATH"))
	if err != nil {
		return err
	}
	cmd := exec.Command("go", "test")

	if verbose {
		cmd.Args = append(cmd.Args, "-v")
	}
	cmd.Dir = path.Join(goPath, "src", importPath)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}
