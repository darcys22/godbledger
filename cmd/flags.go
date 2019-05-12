// Package cmd defines the command line flags for the shared utlities.
package cmd

import (
	"github.com/urfave/cli"
)

var (
	// VerbosityFlag defines the logrus configuration.
	VerbosityFlag = cli.StringFlag{
		Name:  "verbosity",
		Usage: "Logging verbosity (debug, info=default, warn, error, fatal, panic)",
		Value: "info",
	}
	// DataDirFlag defines a path on disk.
	DataDirFlag = DirectoryFlag{
		Name:  "datadir",
		Usage: "Data directory for the databases and keystore",
		Value: DirectoryString{DefaultDataDir()},
	}
	// ClearDB tells the beacon node to remove any previously stored data at the data directory.
	ClearDB = cli.BoolFlag{
		Name:  "clear-db",
		Usage: "Clears any previously stored data at the data directory",
	}
)
