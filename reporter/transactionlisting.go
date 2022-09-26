package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"encoding/csv"
	"encoding/json"

	"github.com/darcys22/godbledger/godbledger/cmd"
	"github.com/darcys22/godbledger/godbledger/ledger"

	"github.com/olekukonko/tablewriter"
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
	Usage: "reporter transactions [(--json | --csv) <output-filename> ]",
	Description: `
Lists all Transactions in the Database

If you want to see all the transactions in the database, or export to CSV/JSON
`,
	Flags: []cli.Flag{
		csvFlag,
		jsonFlag,
		formattingFlag,
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
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Date", "ID", "Account", "Description", "Currency", "Amount"})
		table.SetBorder(false)
		table.SetAlignment(tablewriter.ALIGN_RIGHT)

		queryDateStart := time.Now().Add(time.Hour * 24 * 365 * -100)
		queryDateEnd := time.Now().Add(time.Hour * 24 * 365 * 100)

		queryDB := `
			SELECT
				transactions.transaction_id,
				splits.split_date,
				splits.description,
				splits.currency,
				currency.decimals,
				splits.amount,
				split_accounts.account_id
			FROM
				splits
				JOIN split_accounts ON splits.split_id = split_accounts.split_id
				JOIN transactions on splits.transaction_id = transactions.transaction_id
				JOIN currencies AS currency ON splits.currency = currency.NAME
			WHERE
				splits.split_date >= ?
				AND splits.split_date <= ?
				AND "void" NOT IN(
					SELECT
						t.tag_name
					FROM
						tags AS t
						JOIN transaction_tag AS tt ON tt.tag_id = t.tag_id
					WHERE
						tt.transaction_id = splits.transaction_id
				)
				AND "main" IN (
				  SELECT
					  t.tag_name
					FROM
					  tags AS t
					  JOIN account_tag AS at ON at.tag_id = t.tag_id
					WHERE
						at.account_id = split_accounts.account_id)
		;`

		log.Debug("Querying Database")
		rows, err := ledger.LedgerDb.Query(queryDB, queryDateStart, queryDateEnd)

		if err != nil {
			return fmt.Errorf("Could not query database (%v)", err)
		}
		defer rows.Close()

		for rows.Next() {
			// Scan one customer record
			var t Transaction
			var decimals float64
			if err := rows.Scan(&t.ID, &t.Date, &t.Description, &t.Currency, &decimals, &t.Amount, &t.Account); err != nil {
				return fmt.Errorf("Could not scan rows of query (%v)", err)
			}
			centsAmount, err := strconv.ParseFloat(t.Amount, 64)
			if err != nil {
				return fmt.Errorf("Could not process the amount as a float (%v)", err)
			}
			if ctx.Bool("unformatted") {
				t.Amount = fmt.Sprintf("%.2f", centsAmount/math.Pow(10, decimals))
			} else {
				p := message.NewPrinter(language.English)
				t.Amount = p.Sprintf("$%.2f", centsAmount/math.Pow(10, decimals))
			}
			output.Data = append(output.Data, t)
			table.Append([]string{t.Date, t.ID, t.Account, t.Description, t.Currency, t.Amount})
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
			csvWriter.Write([]string{"Date", "ID", "Account", "Description", "Currency", "Amount"})

			for _, element := range output.Data {
				err := csvWriter.Write([]string{element.Date, element.ID, element.Account, element.Description, element.Currency, element.Amount})
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

			bytes, err := json.Marshal(output.Data)
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
