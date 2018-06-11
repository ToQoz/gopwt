package calc_gopwt_test

import (
	"flag"
	"os"
	"testing"

	"github.com/ToQoz/gopwt"
	"github.com/ToQoz/gopwt/assert"
	. "github.com/bmuschko/go-testing-frameworks/calc"
)

func TestMain(m *testing.M) {
	flag.Parse()
	gopwt.Empower()
	os.Exit(m.Run())
}

func TestAdd(t *testing.T) {
	result := Add(1, 2)
	assert.OK(t, 3 == result)
}
