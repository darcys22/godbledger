package tests

import (
	"flag"
	"testing"

	"github.com/urfave/cli/v2"

	"github.com/darcys22/godbledger/godbledger/cmd"
)

func TestEndToEnd_MinimalConfig(t *testing.T) {

	set := flag.NewFlagSet("test", 0)
	set.String("config", "", "doc")

	ctx := cli.NewContext(nil, set, nil)

	err, cfg := cmd.MakeConfig(ctx)
	if err != nil {
		t.Fatalf("New Config Failed: %v", err)
	}

	cfg.DatabaseType = "memorydb"

	runEndToEndTest(t, cfg)
}
