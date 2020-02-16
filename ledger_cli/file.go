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
	Usage:     "Reads a ledger file",
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
			//ledgerFileName = c.Args().Get(0)

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
