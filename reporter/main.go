package main

import (
	"fmt"
	"os"

	"github.com/darcys22/godbledger/godbledger/cmd"

	"github.com/urfave/cli/v2"

	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

const ()

var app *cli.App

func init() {
	customFormatter := new(prefixed.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.FullTimestamp = true
	logrus.SetFormatter(customFormatter)
	app = cli.NewApp()
	app.Name = "Reporter"
	app.Usage = "Extracts GL and TB reports from a Godbledger database"
	app.Commands = []*cli.Command{
		// transactionlisting.go
		commandTransactionListing,
		// trialbalance.go
		commandTrialBalance,
		// pdfgenerator.go
		commandPDFGenerate,
	}
	app.Flags = []cli.Flag{
		cmd.VerbosityFlag,
		cmd.ConfigFileFlag,
		cmd.RPCHost,
		cmd.RPCPort,
		cmd.CertFlag,
		cmd.KeyFlag,
	}
	app.Action = reporterConsole
}

// Commonly used command line flags.
var (
	csvFlag = &cli.StringFlag{
		Name:  "csv",
		Usage: "output CSV instead of human-readable format",
	}
	jsonFlag = &cli.StringFlag{
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
