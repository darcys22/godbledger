package main

import (
	"io"
	"os/exec"
	"strings"
	//"log"
	"net/http"
	"os"

	//"database/sql"
	//_ "github.com/mattn/go-sqlite3"

	//"github.com/darcys22/godbledger/godbledger/cmd"

	"github.com/urfave/cli"
)

//type Account struct {
//Account string `json:"account"`
//Amount  string `json:"amount"`
//}

//var tboutput struct {
//Data []Account `json:"data"`
//}

var commandPDFGenerate = cli.Command{
	Name:      "pdf",
	Usage:     "Creates a PDF of the Financials",
	ArgsUsage: "[]",
	Description: `
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
		//err, cfg := cmd.MakeConfig(ctx)
		//databasefilepath := ctx.Args().First()
		//if databasefilepath == "" {
		//databasefilepath = cfg.DatabaseLocation
		//}
		//if _, err := os.Stat(databasefilepath); err != nil {
		//panic(fmt.Sprintf("Database does not already exist at %s.", databasefilepath))
		//}

		//SqliteDB, err := sql.Open("sqlite3", databasefilepath)
		//if err != nil {
		//log.Fatal(err)
		//}

		//queryDB := `
		//SELECT
		//split_accounts.account_id,
		//SUM(splits.amount)
		//FROM splits
		//JOIN split_accounts
		//ON splits.split_id = split_accounts.split_id
		//GROUP  BY split_accounts.account_id
		//;`

		//rows, err := SqliteDB.Query(queryDB)
		//if err != nil {
		//log.Fatal(err)
		//}
		//defer rows.Close()

		//for rows.Next() {
		//Scan one customer record
		//var t Account
		//if err := rows.Scan(&t.Account, &t.Amount); err != nil {
		//handle error
		//}
		//tboutput.Data = append(tboutput.Data, t)
		//table.Append([]string{t.Account, t.Amount})
		//}
		//if rows.Err() != nil {
		//handle error
		//}

		dir := "src"

		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				panic(err)
			}
		}

		if err := DownloadFile("./src/financials.html", "https://raw.githubusercontent.com/darcys22/pdf-generator/master/financials.html"); err != nil {
			panic(err)
		}

		if err := DownloadFile("./src/pdfgenerator.js", "https://raw.githubusercontent.com/darcys22/pdf-generator/master/pdfgenerator.js"); err != nil {
			panic(err)
		}

		if err := DownloadFile("./src/data.json", "https://raw.githubusercontent.com/darcys22/pdf-generator/master/data.json"); err != nil {
			panic(err)
		}

		command := "node ./pdfgenerator.js"
		parts := strings.Fields(command)
		cmd := exec.Command(parts[0], parts[1:]...)
		cmd.Dir = "./src"

		cmd.Run()
		err := os.Rename("src/mypdf.pdf", "financials.pdf")
		if err != nil {
			panic(err)
		}
		err = os.RemoveAll("src")
		if err != nil {
			panic(err)
		}

		return nil
	},
}

func DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
