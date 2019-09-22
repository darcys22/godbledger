package main

import (
	"fmt"
	"log"
	"os"
	"os/user"

	"github.com/urfave/cli"
)

const ()

var app *cli.App

func init() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	defaultDBName = usr.HomeDir + "/.ledger/ledgerdata/ledger.db"
	app = cli.NewApp()
	app.Name = "Reporter"
	app.Usage = "Provides GL and TB reports for GoDBLedger"
	app.Commands = []cli.Command{
		commandTransactionListing,
		commandTrialBalance,
	}
	app.Action = reporterConsole
}

// Commonly used command line flags.
var (
	defaultDBName = "/.ledger/ledgerdata/ledger.db"
	csvFlag       = cli.BoolFlag{
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
