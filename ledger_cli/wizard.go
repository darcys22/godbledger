package main

import (
	"context"
	"log"
	"math/big"
	"time"

	"github.com/urfave/cli"
)

var commandWizardJournal = cli.Command{
	Name:      "single",
	Usage:     "creates and submits a single transaction",
	ArgsUsage: "[]",
	Description: `
`,
	Flags: []cli.Flag{},
	Action: func(c *cli.Context) error {

		reader := bufio.NewReader(os.Stdin)
		fmt.Println(text)

		fmt.Println("Journal Entry Wizard")
		fmt.Println("--------------------")

		fmt.Print("Enter the date (yyyy-mm-dd): ")
		datetext, _ := reader.ReadString('\n')
		date, _ := time.Parse("2006-01-02", datetext)

		fmt.Print("Enter the Journal Descritpion: ")
		desc, _ := reader.ReadString('\n')

		transactionLines := make([]Account, 2)

		line1Account := "Expenses:Groceries"
		line1Desc := "Groceries"
		line1Amount := big.NewRat(7500, 1)

		transactionLines[0] = Account{
			Name:        line1Account,
			Description: line1Desc,
			Balance:     line1Amount,
		}

		line2Account := "Assets:Checking"
		line2Desc := "Groceries"
		line2Amount := big.NewRat(-7500, 1)

		transactionLines[1] = Account{
			Name:        line2Account,
			Description: line2Desc,
			Balance:     line2Amount,
		}

		req := &Transaction{
			Date:           date,
			Payee:          desc,
			AccountChanges: transactionLines,
			Signature:      "stuff",
		}

		err := Send(req)
		if err != nil {
			log.Fatalf("could not send: %v", err)
		}

		return nil
	},
}
