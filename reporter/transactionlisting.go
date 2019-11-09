package main

import (
	"fmt"
	"log"
	"os"

	"database/sql"
	"encoding/csv"
	_ "github.com/mattn/go-sqlite3"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
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

var commandTransactionListing = cli.Command{
	Name:      "transactions",
	Usage:     "List all Transactions in the Database",
	ArgsUsage: "[]",
	Description: `
Lists all Transactions in the Database

If you want to see all the transactions in the database, or export to CSV
`,
	Flags: []cli.Flag{
		csvFlag,
		//cli.StringFlag{
		//Name:  "privatekey",
		//Usage: "file containing a raw private key to encrypt",
		//},
	},
	Action: func(ctx *cli.Context) error {
		//Check if keyfile path given and make sure it doesn't already exist.
		databasefilepath := ctx.Args().First()
		if databasefilepath == "" {
			databasefilepath = defaultDBName
		}
		if _, err := os.Stat(databasefilepath); err != nil {
			panic(fmt.Sprintf("Database does not already exist at %s.", databasefilepath))
		}

		SqliteDB, err := sql.Open("sqlite3", databasefilepath)
		if err != nil {
			log.Fatal(err)
		}
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

		rows, err := SqliteDB.Query(queryDB)
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
		if rows.Err() != nil {
			// handle error
		}

		//Output some information.
		if ctx.Bool(csvFlag.Name) {
			file, err := os.OpenFile("test.csv", os.O_CREATE|os.O_WRONLY, 0777)
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

			fmt.Println("CSV Yo")
		} else {
			fmt.Println()
			table.Render()
			fmt.Println()
		}
		return nil
	},
}
