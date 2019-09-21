package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
)

const (
	defaultDBName = "~/.ledger/ledgerdata/ledger.db"
)

var app *cli.App

func init() {
	app := cli.NewApp()
	app.Name = "Reporter"
	app.Usage = "Provides GL and TB reports for GoDBLedger"
	app.Commands = []cli.Command{
		commandTransactionListing,
	}
}

// Commonly used command line flags.
var (
	csvFlag = cli.BoolFlag{
		Name:  "csv",
		Usage: "output CSV instead of human-readable format",
	}
)

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
