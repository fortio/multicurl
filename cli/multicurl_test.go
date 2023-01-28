package cli_test

import (
	"os"
	"testing"

	"github.com/fortio/multicurl/cli"
	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"multicurl": cli.Main,
	}))
}

func TestMulticurl(t *testing.T) {
	testscript.Run(t, testscript.Params{Dir: "./"})
}
