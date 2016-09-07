package gopwt

import (
	"flag"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	flag.Parse()
	Empower()
	os.Exit(m.Run())
}
