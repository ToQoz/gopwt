package gopwt

import (
	"flag"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	flag.Parse()
	Main()
	os.Exit(m.Run())
}
