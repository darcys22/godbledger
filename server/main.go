package main

import (
	//"fmt"
	"context"
	"os"
	//"strconv"

	"github.com/darcys22/godbledger/server/cmd"
	"github.com/darcys22/godbledger/server/ledger"
	"github.com/darcys22/godbledger/server/node"
	"github.com/darcys22/godbledger/server/rpc"
	"github.com/darcys22/godbledger/server/version"

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

	fullnode, err := node.New(ctx)
	if err != nil {
		return err
	}
	ledger, err := ledger.New(ctx)
	fullnode.Register(ledger)
	rpc := rpc.NewRPCService(context.Background(), &rpc.Config{Port: "50051"}, ledger)
	fullnode.Register(rpc)
	fullnode.Start()

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
	app.Usage = "Accounting Ledger Database Server"
	app.Action = startNode
	app.Version = version.Version
	app.Commands = []cli.Command{
		// See config.go
		cmd.DumpConfigCommand,
	}

	app.Flags = []cli.Flag{
		cmd.VerbosityFlag,
		cmd.DataDirFlag,
		cmd.ClearDB,
	}

	if err := app.Run(os.Args); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

}
