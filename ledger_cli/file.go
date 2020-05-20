//Copyright (c) 2013 Chris Howey

//Permission to use, copy, modify, and distribute this software for any
//purpose with or without fee is hereby granted, provided that the above
//copyright notice and this permission notice appear in all copies.

//THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
//WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
//MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
//ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
//WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
//ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
//OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

package main

import (
	//"flag"
	"fmt"
	"log"
	"strings"
	"unicode/utf8"

	"github.com/urfave/cli"
)

const (
	transactionDateFormat = "2006/01/02"
	displayPrecision      = 2
)

var commandFile = &cli.Command{
	Name:      "file",
	Usage:     "ledger_cli file ./test/transaction-codes-1.test",
	ArgsUsage: "[]",
	Description: `
	Loads a file in the ledger cli format
`,
	Flags: []cli.Flag{},
	Action: func(ctx *cli.Context) error {

		var ledgerFileName string

		if ctx.NArg() > 0 {
			columnWidth := 80

			ledgerFileName = "test/transaction-codes-2.test"
			if len(ctx.Args().Get(0)) > 0 {
				ledgerFileName = ctx.Args().Get(0)
			}

			ledgerFileReader, err := NewLedgerReader(ledgerFileName)
			if err != nil {
				log.Printf("error reading file, %v\n", err)
				return err
			}

			generalLedger, parseError := ParseLedger(ledgerFileReader)
			if parseError != nil {
				log.Printf("error parsing file, %s\n", parseError.Error())
				return parseError
			}

			PrintLedger(generalLedger, columnWidth)
			SendLedger(generalLedger)
		} else {
			log.Printf("This command requires an argument.")
		}
		return nil
	},
}

// PrintTransaction prints a transaction formatted to fit in specified column width.
func PrintTransaction(trans *Transaction, columns int) {
	fmt.Printf("%+v\n", trans)
	fmt.Printf("%s %s\n", trans.Date.Format(transactionDateFormat), trans.Payee)
	for _, accChange := range trans.AccountChanges {
		outBalanceString := accChange.Balance.FloatString(displayPrecision)
		spaceCount := columns - 4 - utf8.RuneCountInString(accChange.Name) - utf8.RuneCountInString(outBalanceString)
		fmt.Printf("    %s%s%s\n", accChange.Name, strings.Repeat(" ", spaceCount), outBalanceString)
	}
	fmt.Println("")
}

// PrintLedger prints all transactions as a formatted ledger file.
func PrintLedger(generalLedger []*Transaction, columns int) {
	for _, trans := range generalLedger {
		PrintTransaction(trans, columns)
	}
}

func SendLedger(generalLedger []*Transaction) {
	for _, trans := range generalLedger {
		Send(trans)
	}
}
