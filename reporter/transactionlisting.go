package main

import (
	"fmt"
	//"os"
	//"path/filepath"

	"github.com/urfave/cli"
)

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
		fmt.Println(databasefilepath)
		//if _, err := os.Stat(databasefilepath); err != nil {
		//panic(fmt.Sprintf("Database does not already exist at %s.", databasefilepath))
		//}

		//fmt.Println("Success")

		// Output some information.
		//out := outputGenerate{
		//Address: key.Address.Hex(),
		//}
		//if ctx.Bool(jsonFlag.Name) {
		//mustPrintJSON(out)
		//} else {
		//fmt.Println("Address:", out.Address)
		//}
		return nil
	},
}
