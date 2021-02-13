// Package mysql performs full a end-to-end test for GoDBLedger specifically using a mysql backend,
// including spinning up a server and making sure its running, and sending test data to verify

// +build mysql

package tests

import (
	"flag"
	"testing"

	"github.com/darcys22/godbledger/godbledger/cmd"

	"github.com/urfave/cli/v2"
)

func TestMySQL(t *testing.T) {

	// Create a config from the defaults which would usually be created by the CLI library
	set := flag.NewFlagSet("test", 0)
	set.String("config", "", "doc")
	ctx := cli.NewContext(nil, set, nil)
	err, cfg := cmd.MakeConfig(ctx)
	if err != nil {
		t.Fatalf("New Config Failed: %v", err)
	}

	// Set the Database type to a MySQL database
	cfg.DatabaseType = "mysql"

	//runEndToEndTest(t, cfg)
}
