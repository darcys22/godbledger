package tests

import (
	"flag"
	"testing"

	"github.com/urfave/cli/v2"

	"github.com/darcys22/godbledger/godbledger/cmd"

	e2eParams "github.com/darcys22/godbledger/tests/params"
)

func TestEndToEnd_MinimalConfig(t *testing.T) {

	// Create a config from the defaults which would usually be created by the CLI library
	set := flag.NewFlagSet("test", 0)
	set.String("config", "", "doc")
	ctx := cli.NewContext(nil, set, nil)
	err, cfg := cmd.MakeConfig(ctx)
	if err != nil {
		t.Fatalf("New Config Failed: %v", err)
	}

	// Set the Database type to a SQLite3 in memory database
	cfg.DatabaseType = "memorydb"

	// Initialises Logpath etc
	if err := e2eParams.Init(); err != nil {
		t.Fatal(err)
	}
	runEndToEndTest(t, cfg)
}
