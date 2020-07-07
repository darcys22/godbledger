package main

import (
	"fmt"
	"os"

	"encoding/csv"
	"encoding/json"

	"github.com/darcys22/godbledger/godbledger/cmd"
	"github.com/darcys22/godbledger/godbledger/ledger"

	"github.com/olekukonko/tablewriter"
	//"github.com/urfave/cli"
	"github.com/urfave/cli/v2"
)

type Transaction struct {
	Account     string `json:"account"`
	ID          string `json:"id"`
	Date        string `json:"date"`
	Description string `json:"desc"`
	Amount      string `json:"amount"`
	Currency    string `json:"currency"`
}

var output struct {
	Data []Transaction `json:"data"`
}

var commandTransactionListing = &cli.Command{
	Name:  "transactions",
	Usage: "ledger_cli transactions [(--json | --csv) <output-filename> ]",
	Description: `
Lists all Transactions in the Database

If you want to see all the transactions in the database, or export to CSV/JSON
`,
	Flags: []cli.Flag{
		csvFlag,
		jsonFlag,
	},
	Action: func(ctx *cli.Context) error {
		//Check if keyfile path given and make sure it doesn't already exist.

		err, cfg := cmd.MakeConfig(ctx)
		databasefilepath := ctx.Args().First()
		if databasefilepath == "" {
			databasefilepath = cfg.DatabaseLocation
		}
		ledger, err := ledger.New(ctx, cfg)
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Date", "ID", "Account", "Description", "Currency", "Amount"})
		table.SetBorder(false)

		queryDB := `
			SELECT 
				transactions.transaction_id,
				splits.split_date,
				splits.description,
				splits.currency,
				splits.amount,
				split_accounts.account_id
			FROM splits 
				JOIN split_accounts 
					ON splits.split_id = split_accounts.split_id
				JOIN transactions
					on splits.transaction_id = transactions.transaction_id
		;`

		log.Debug("Querying Database")
		rows, err := ledger.LedgerDb.Query(queryDB)

		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		for rows.Next() {
			// Scan one customer record
			var t Transaction
			if err := rows.Scan(&t.ID, &t.Date, &t.Description, &t.Currency, &t.Amount, &t.Account); err != nil {
				// handle error
			}
			output.Data = append(output.Data, t)
			table.Append([]string{t.Date, t.ID, t.Account, t.Description, t.Currency, t.Amount})
		}
		fmt.Printf("Database does not already exist at %s.\n", databasefilepath)
		if rows.Err() != nil {
			// handle error
		}

		//Output some information.
		if len(ctx.String(csvFlag.Name)) > 0 {
			log.Infof("Exporting CSV to %s", ctx.String(csvFlag.Name))

			file, err := os.OpenFile(ctx.String(csvFlag.Name), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
			defer file.Close()

			if err != nil {
				os.Exit(1)
			}

			csvWriter := csv.NewWriter(file)
			defer csvWriter.Flush()
			csvWriter.Write([]string{"Date", "ID", "Account", "Description", "Currency", "Amount"})

			for _, element := range output.Data {
				err := csvWriter.Write([]string{element.Date, element.ID, element.Account, element.Description, element.Currency, element.Amount})
				if err != nil {
					log.Fatal("Cannot write to file", err)
				}
			}

		} else if len(ctx.String(jsonFlag.Name)) > 0 {
			log.Infof("Exporting JSON to %s", ctx.String(jsonFlag.Name))
			file, err := os.OpenFile(ctx.String(jsonFlag.Name), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)

			if err != nil {
				log.Fatalf("Cannot open to file: %s", err)
			}
			defer file.Close()

			bytes, err := json.Marshal(output.Data)
			if err != nil {
				log.Fatal("Cannot serialize")
			}
			_, err = file.Write(bytes)
			if err != nil {
				log.Fatal("Cannot write to file", err)
			}

		} else {
			fmt.Println()
			table.Render()
			fmt.Println()
		}
		return nil
	},
}
