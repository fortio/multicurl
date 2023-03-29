package mc_test

import (
	"context"
	"os"
	"testing"

	"fortio.org/multicurl/cli"
	"fortio.org/multicurl/mc"
	"fortio.org/testscript"
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
	a, err := mc.ResolveAll(context.Background(), "[::1]", "ip")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(a) != 1 {
		t.Errorf("Unexpected number of addresses: %d", len(a))
	}
	aStr := a[0].String()
	if aStr != "::1" {
		t.Errorf("Unexpected address: %s", aStr)
	}
}
