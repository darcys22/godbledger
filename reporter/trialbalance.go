package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"encoding/csv"
	"encoding/json"

	"github.com/darcys22/godbledger/godbledger/cmd"
	"github.com/darcys22/godbledger/godbledger/ledger"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"
)

type Account struct {
	Account string `json:"account"`
	Amount  string `json:"amount"`
}

var tboutput struct {
	Data []Account `json:"data"`
}

var commandTrialBalance = &cli.Command{
	Name:  "trialbalance",
	Usage: "reporter trialbalance [(--json | --csv) <output-filename> ]",
	Description: `
Sums the value of the transactions per account in the Database

If you want to see all the transactions in the database, or export to CSV
`,
	Flags: []cli.Flag{
		csvFlag,
		jsonFlag,
	},
	Action: func(ctx *cli.Context) error {
		//Check if keyfile path given and make sure it doesn't already exist.
		err, cfg := cmd.MakeConfig(ctx)
		if err != nil {
			return fmt.Errorf("Could not make config (%v)", err)
		}

		ledger, err := ledger.New(ctx, cfg)
		if err != nil {
			return fmt.Errorf("Could not make new ledger (%v)", err)
		}
		queryDate := time.Now()

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Account", fmt.Sprintf("Balance at %s", queryDate.Format("02 January 2006"))})
		table.SetBorder(false)

		queryDB := `
			SELECT split_accounts.account_id,
						 Sum(splits.amount),
						 currency.decimals
			FROM   splits
						 JOIN split_accounts ON splits.split_id = split_accounts.split_id
						 JOIN currencies AS currency ON splits.currency = currency.NAME
			WHERE  splits.split_date <= ?
						 AND "void" NOT IN (SELECT t.tag_name
																FROM   tags AS t
																			 JOIN transaction_tag AS tt
																				 ON tt.tag_id = t.tag_id
																WHERE  tt.transaction_id = splits.transaction_id)
						 AND "main" IN (SELECT t.tag_name
																FROM   tags AS t
																			 JOIN account_tag AS at
																				 ON at.tag_id = t.tag_id
																WHERE  at.account_id = split_accounts.account_id)
			GROUP  BY split_accounts.account_id, splits.currency

			;`

		log.Debug("Querying Database")
		rows, err := ledger.LedgerDb.Query(queryDB, queryDate)
		if err != nil {
			return fmt.Errorf("Could not query database (%v)", err)
		}
		defer rows.Close()

		for rows.Next() {
			// Scan one customer record
			var t Account
			var decimals float64
			if err := rows.Scan(&t.Account, &t.Amount, &decimals); err != nil {
				return fmt.Errorf("Could not scan rows of query (%v)", err)
			}
			centsAmount, err := strconv.ParseFloat(t.Amount, 64)
			if err != nil {
				return fmt.Errorf("Could not process the amount as a float (%v)", err)
			}
			t.Amount = fmt.Sprintf("%.2f", centsAmount/math.Pow(10, decimals))
			tboutput.Data = append(tboutput.Data, t)
			table.Append([]string{t.Account, t.Amount})
		}
		if rows.Err() != nil {
			return fmt.Errorf("rows errored with (%v)", rows.Err())
		}

		//Output some information.
		if len(ctx.String(csvFlag.Name)) > 0 {
			log.Infof("Exporting CSV to %s", ctx.String(csvFlag.Name))
			file, err := os.OpenFile(ctx.String(csvFlag.Name), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				return fmt.Errorf("opening csv file errored with (%v)", err)
			}
			defer file.Close()

			csvWriter := csv.NewWriter(file)
			defer csvWriter.Flush()
			csvWriter.Write([]string{"Account", "Balance"})

			for _, element := range tboutput.Data {
				err := csvWriter.Write([]string{element.Account, element.Amount})
				if err != nil {
					return fmt.Errorf("could not write to csv file (%v)", err)
				}
			}

		} else if len(ctx.String(jsonFlag.Name)) > 0 {
			log.Infof("Exporting JSON to %s", ctx.String(jsonFlag.Name))
			file, err := os.OpenFile(ctx.String(jsonFlag.Name), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)

			if err != nil {
				return fmt.Errorf("could not open json file (%v)", err)
			}
			defer file.Close()

			bytes, err := json.Marshal(tboutput.Data)
			if err != nil {
				return fmt.Errorf("could not serialise json (%v)", err)
			}
			_, err = file.Write(bytes)
			if err != nil {
				return fmt.Errorf("could not write to json file (%v)", err)
			}
		} else {
			fmt.Println()
			table.Render()
			fmt.Println()
		}
		return nil
	},
}
