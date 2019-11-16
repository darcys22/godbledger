// Package cmd defines the command line flags for the shared utlities.
package cmd

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	"github.com/urfave/cli"
)

func homeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}

// DefaultDataDir is the default data directory to use for the databases and other
// persistence requirements.
func DefaultDataDir() string {
	// Try to place the data folder in the user's home dir
	home := homeDir()
	if home != "" {
		if runtime.GOOS == "darwin" {
			return filepath.Join(home, "Library", "ledger")
		} else if runtime.GOOS == "windows" {
			return filepath.Join(home, "AppData", "Roaming", "ledger")
		} else {
			return filepath.Join(home, ".ledger")
		}
	}
	// As we cannot guess a stable location, return empty and handle later
	return ""
}

var (
	// VerbosityFlag defines the logrus configuration.
	VerbosityFlag = cli.StringFlag{
		Name:  "verbosity",
		Usage: "Logging verbosity (debug, info=default, warn, error, fatal, panic)",
		//Value: "debug",
	}
	// DataDirFlag defines a path on disk.
	DataDirFlag = cli.StringFlag{
		Name:  "datadir",
		Usage: "Data directory for the databases and keystore",
		//Value: DefaultDataDir(),
	}
	// ClearDB tells the node to remove any previously stored data at the data directory.
	ClearDB = cli.BoolFlag{
		Name:  "clear-db",
		Usage: "Clears any previously stored data at the data directory",
	}
	ConfigFileFlag = cli.StringFlag{
		Name:  "config",
		Usage: "TOML configuration file",
	}
)
