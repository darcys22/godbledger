package main

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"

	"github.com/darcys22/godbledger/godbledger/cmd"
	"github.com/darcys22/godbledger/godbledger/ledger"
	"github.com/darcys22/godbledger/godbledger/node"
	"github.com/darcys22/godbledger/godbledger/rpc"
	"github.com/darcys22/godbledger/godbledger/version"
)

func startNode(ctx *cli.Context) error {
	err, cfg := cmd.MakeConfig(ctx)
	if err != nil {
		return err
	}

	fullnode, err := node.New(ctx)
	if err != nil {
		return err
	}
	ledger, err := ledger.New(ctx, cfg)
	if err != nil {
		return err
	}
	fullnode.Register(ledger)
	rpc := rpc.NewRPCService(context.Background(), &rpc.Config{
		Host:     cfg.Host,
		Port:     cfg.RPCPort,
		CertFlag: cfg.Cert,
		KeyFlag:  cfg.Key,
	}, ledger)
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
	app.Name = "Godbledger"
	app.Usage = "Accounting Ledger Database Server"
	app.EnableBashCompletion = true
	app.Action = startNode
	app.Version = version.Version
	app.Commands = []*cli.Command{
		// See config.go
		cmd.DumpConfigCommand,
		cmd.InitConfigCommand,
	}

	app.Flags = []cli.Flag{
		cmd.VerbosityFlag,
		cmd.DataDirFlag,
		cmd.ClearDB,
		cmd.ConfigFileFlag,
		cmd.RPCHost,
		cmd.RPCPort,
		cmd.CertFlag,
		cmd.KeyFlag,
	}

	if err := app.Run(os.Args); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

}
