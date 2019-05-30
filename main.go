package main

import (
	//"fmt"
	"os"
	//"strconv"

	"godbledger/cmd"
	"godbledger/node"
	"godbledger/version"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

func startNode(ctx *cli.Context) error {
	verbosity := ctx.GlobalString(cmd.VerbosityFlag.Name)
	level, err := logrus.ParseLevel(verbosity)
	if err != nil {
		return err
	}
	logrus.SetLevel(level)

	ledger, err := node.New(ctx)
	if err != nil {
		return err
	}
	ledger.Start()

	//statement, _ := ledger.DB.DB.Prepare("CREATE TABLE IF NOT EXISTS people (id INTEGER PRIMARY KEY, firstname TEXT, lastname TEXT)")
	//statement.Exec()
	//statement, _ = ledger.DB.DB.Prepare("INSERT INTO people (firstname, lastname) VALUES (?, ?)")
	//statement.Exec("Nic", "Raboy")
	//rows, _ := ledger.DB.DB.Query("SELECT id, firstname, lastname FROM people")

	//var id int
	//var firstname string
	//var lastname string
	//for rows.Next() {
	//rows.Scan(&id, &firstname, &lastname)
	//fmt.Println(strconv.Itoa(id) + ": " + firstname + " " + lastname)
	//}

	return nil
}

func main() {
	customFormatter := new(prefixed.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.FullTimestamp = true
	logrus.SetFormatter(customFormatter)
	log := logrus.WithField("prefix", "main")
	app := cli.NewApp()
	app.Name = "ledger"
	app.Usage = "Accounting Ledger Database for the 21st Century"
	app.Action = startNode
	app.Version = version.Version

	app.Flags = []cli.Flag{
		cmd.VerbosityFlag,
		cmd.ClearDB,
	}

	if err := app.Run(os.Args); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

}
