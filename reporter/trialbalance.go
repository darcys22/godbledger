package main

import (
	"fmt"
	"os"

	"database/sql"
	"encoding/csv"
	_ "github.com/mattn/go-sqlite3"

	"github.com/darcys22/godbledger/godbledger/cmd"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
)

type Account struct {
	Account string `json:"account"`
	Amount  string `json:"amount"`
}

var tboutput struct {
	Data []Account `json:"data"`
}

var commandTrialBalance = &cli.Command{
	Name:      "trialbalance",
	Usage:     "List all Transactions in the Database",
	ArgsUsage: "[]",
	Description: `
Sums the value of the transactions per account in the Database

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
		err, cfg := cmd.MakeConfig(ctx)
		databasefilepath := ctx.Args().First()
		if databasefilepath == "" {
			databasefilepath = cfg.DatabaseLocation
		}
		if _, err := os.Stat(databasefilepath); err != nil {
			panic(fmt.Sprintf("Database does not already exist at %s.", databasefilepath))
		}

		DB, err := sql.Open(cfg.DatabaseType, databasefilepath)
		if err != nil {
			log.Fatal(err)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Account", "Balance"})
		table.SetBorder(false)

		queryDB := `
			SELECT 
				split_accounts.account_id,
				SUM(splits.amount)
			FROM splits 
				JOIN split_accounts 
				ON splits.split_id = split_accounts.split_id
			GROUP  BY split_accounts.account_id
		;`

		rows, err := DB.Query(queryDB)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		for rows.Next() {
			// Scan one customer record
			var t Account
			if err := rows.Scan(&t.Account, &t.Amount); err != nil {
				// handle error
			}
			tboutput.Data = append(tboutput.Data, t)
			table.Append([]string{t.Account, t.Amount})
		}
		if rows.Err() != nil {
			// handle error
		}

		//Output some information.
		if len(ctx.String(csvFlag.Name)) > 0 {
			file, err := os.OpenFile(ctx.String(csvFlag.Name), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
			defer file.Close()

			if err != nil {
				os.Exit(1)
			}

			csvWriter := csv.NewWriter(file)
			defer csvWriter.Flush()
			csvWriter.Write([]string{"Account", "Balance"})

			for _, element := range tboutput.Data {
				err := csvWriter.Write([]string{element.Account, element.Amount})
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
		//if ctx.Bool(jsonFlag.Name) {
		//mustPrintJSON(out)
		//} else {
		//fmt.Println("Address:", out.Address)
		//}
		return nil
	},
}
