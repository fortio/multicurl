package mc_test

import (
	"context"
	"os"
	"testing"

	"github.com/fortio/multicurl/cli"
	"github.com/fortio/multicurl/mc"
	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"multicurl": cli.Main,
	}))
}

func TestMulticurl(t *testing.T) {
	testscript.Run(t, testscript.Params{Dir: "../cli/"})
}

func TestResolveIP6(t *testing.T) {
	p, a, err := mc.ResolveAll(context.Background(), "[::1]", "https", "ip")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if p != 443 {
		t.Errorf("Unexpected port: %d", p)
	}
	if len(a) != 1 {
		t.Errorf("Unexpected number of addresses: %d", len(a))
	}
	aStr := a[0].String()
	if aStr != "::1" {
		t.Errorf("Unexpected address: %s", aStr)
	}
}
