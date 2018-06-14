package issue44

import (
	"flag"
	"os"

	"github.com/ToQoz/gopwt"
	"github.com/ToQoz/gopwt/assert"

	"testing"
)

func TestMain(m *testing.M) {
	flag.Parse()
	gopwt.Empower()
	os.Exit(m.Run())
}

func TestFoo(t *testing.T) {
	assert.OK(t, 1 == 1)
}
