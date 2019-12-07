package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
)

const ()

var app *cli.App

func init() {
	app = cli.NewApp()
	app.Name = "Reporter"
	app.Usage = "Provides GL and TB reports for GoDBLedger"
	app.Commands = []cli.Command{
		commandTransactionListing,
		commandTrialBalance,
		commandPDFGenerate,
	}
	app.Action = reporterConsole
}

// Commonly used command line flags.
var (
	csvFlag = cli.StringFlag{
		Name:  "csv",
		Usage: "output CSV instead of human-readable format",
	}
	jsonFlag = cli.StringFlag{
		Name:  "json",
		Usage: "output json instead of human-readable format",
	}
)

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
