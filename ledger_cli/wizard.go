package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/darcys22/godbledger/godbledger/cmd"

	"github.com/marcmak/calc/calc"
	"github.com/urfave/cli/v2"
)

var commandWizardJournal = &cli.Command{
	Name:      "journal",
	Usage:     "creates and submits a single transaction",
	ArgsUsage: "[]",
	Description: `
`,
	Flags: []cli.Flag{},
	Action: func(ctx *cli.Context) error {
		err, cfg := cmd.MakeConfig(ctx)
		if err != nil {
			return fmt.Errorf("Could not make config (%v)", err)
		}

		reader := bufio.NewReader(os.Stdin)

		fmt.Println("Journal Entry Wizard")
		fmt.Println("--------------------")

		fmt.Print("Enter the date (yyyy-mm-dd): ")
		datetext, _ := reader.ReadString('\n')
		date, err := time.Parse("2006-01-02", strings.TrimSpace(datetext))
		if err != nil {
			return fmt.Errorf("Could not make parse date string %s with error (%v)", datetext, err)
		}

		fmt.Print("Enter the Journal Descripion: ")
		desc, _ := reader.ReadString('\n')
		fmt.Println("")

		count := 0

		var transactionLines []Account

		for {
			count += 1
			fmt.Printf("Line item #%d\n", count)
			fmt.Print("Enter the line Descripion: ")
			lineDesc, _ := reader.ReadString('\n')

			fmt.Print("Enter the Account: ")
			lineAccount, _ := reader.ReadString('\n')

			fmt.Print("Enter the Amount: ")
			lineAmountStr, _ := reader.ReadString('\n')

			lineAmount := new(big.Rat)
			lineAmount.SetFloat64(calc.Solve(lineAmountStr))

			transactionLines = append(transactionLines, Account{
				Name:        lineAccount,
				Description: lineDesc,
				Balance:     lineAmount,
				Currency:    "USD",
			})

			fmt.Print("Would you like to enter more line items? (n to stop): ")
			exit, _ := reader.ReadString('\n')
			fmt.Println("")
			if strings.ToLower(strings.TrimSpace(exit)) == "n" {
				fmt.Println("")
				break
			}
		}

		req := &Transaction{
			Date:           date,
			Payee:          desc,
			AccountChanges: transactionLines,
			Signature:      "stuff",
		}

		bytes, err := json.Marshal(req)
		if err != nil {
			return fmt.Errorf("Can't Serialize Transaction (%v)", err)
		}
		log.Debugf("Transaction: %v => %v, '%v'\n", req, bytes, string(bytes))

		err = Send(cfg, req)
		if err != nil {
			return fmt.Errorf("Could not send transaction (%v)", err)
		}

		return nil
	},
}
