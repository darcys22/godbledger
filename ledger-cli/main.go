/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"

	"github.com/darcys22/godbledger/godbledger/cmd"
)

var log logrus.FieldLogger

var app *cli.App

func init() {
	customFormatter := new(prefixed.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.FullTimestamp = true
	logrus.SetFormatter(customFormatter)
	log = logrus.WithField("prefix", "ledger-cli")
	app = cli.NewApp()
	app.Name = "Ledger CLI"
	app.Usage = "Command Line for GoDBLedger gRPC"
	app.Commands = []*cli.Command{
		// transaction.go
		commandSingleTestTransaction,
		// wizard.go
		commandWizardJournal,
		// jsonjournal.go
		commandJSONJournal,
		// file.go
		commandFile,
		// delete.go
		commandDeleteTransaction,
		commandVoidTransaction,
		// tagaccount.go
		commandTagAccount,
		// addcurrency.go
		commandAddCurrency,
    // addfeedaccount.go
		commandAddFeedAccount,
	}
	app.Flags = []cli.Flag{
		cmd.VerbosityFlag,
		cmd.ConfigFileFlag,
		cmd.RPCHost,
		cmd.RPCPort,
		cmd.CACertFlag,
		cmd.CertFlag,
		cmd.KeyFlag,
	}
	//app.Action = transaction
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
