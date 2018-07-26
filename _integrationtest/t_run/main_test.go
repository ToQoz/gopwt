package t_run

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
	t.Run("t-run", func(t *testing.T) {
		assert.OK(t, 1 == 1)
		t.Run("nest", func(t *testing.T) {
			assert.OK(t, 1 == 1)
		})
	})

	t.Run("t-run2", func(nt *testing.T) {
		assert.OK(nt, 1 == 1)
		nt.Run("nest", func(t *testing.T) {
			assert.OK(t, 1 == 1)
		})
	})

	runHelper(t, "t-run-helper", func(nt *testing.T) {
		assert.OK(nt, 1 == 1)
		runHelper(nt, "nest", func(t *testing.T) {
			assert.OK(t, 1 == 1)
		})
	})
}

func runHelper(t *testing.T, group string, fn func(t *testing.T)) {
	t.Run(group, fn)
}
