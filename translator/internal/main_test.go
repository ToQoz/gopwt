package internal_test

import (
	"flag"
	"os"
	"testing"

	"github.com/ToQoz/gopwt"
)

func TestMain(m *testing.M) {
	flag.Parse()
	gopwt.Empower()
	os.Exit(m.Run())
}
