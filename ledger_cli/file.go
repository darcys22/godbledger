package main

import (
	//"flag"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/urfave/cli"
)

const (
	transactionDateFormat = "2006/01/02"
	displayPrecision      = 2
)

var commandFile = cli.Command{
	Name:      "file",
	Usage:     "Reads a ledger file",
	ArgsUsage: "[]",
	Description: `
Loads a file in the ledger cli format
`,
	Flags: []cli.Flag{},
	Action: func(ctx *cli.Context) error {

		var ledgerFileName string

		ledgerFileName = "test/transaction-codes-2.test"

		columnWidth := 80

		//if len(ledgerFileName) == 0 {
		//flag.Usage()
		//return nil
		//}

		ledgerFileReader, err := NewLedgerReader(ledgerFileName)
		if err != nil {
			fmt.Println(err)
			return err
		}

		generalLedger, parseError := ParseLedger(ledgerFileReader)
		if parseError != nil {
			fmt.Printf("%s\n", parseError.Error())
			return parseError
		}

		PrintLedger(generalLedger, columnWidth)
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
