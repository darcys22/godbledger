package tests

import (
	"flag"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/urfave/cli/v2"

	"github.com/darcys22/godbledger/godbledger/cmd"

	e2eParams "github.com/darcys22/godbledger/tests/params"
)

func TestEndToEnd_MinimalConfig(t *testing.T) {

	// Create a config from the defaults which would usually be created by the CLI library
	set := flag.NewFlagSet("test", 0)
	set.String("datadir", "", "data directory")
	set.String("config", "", "config file")
	set.String("database-location", "", "database location")

	testDataDir, err := filepath.Abs("../release/test-data")
	if err != nil {
		t.Fatalf("Calculating path to test data dir: %v", err)
	}
	testConfigFile := path.Join(testDataDir, "config.toml")
	if fileExists(testConfigFile) {
		err := os.Remove(testConfigFile)
		if err != nil {
			t.Fatalf("Removing old test Config file failed: %v", err)
		}
	}
	testDatabaseLocation := path.Join(testDataDir, "ledger.db")
	if fileExists(testDatabaseLocation) {
		err := os.Remove(testDatabaseLocation)
		if err != nil {
			t.Fatalf("Removing old test database file failed: %v", err)
		}
	}

	set.Set("datadir", testDataDir)
	set.Set("config", testConfigFile)
	set.Set("config", testDatabaseLocation)

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

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
