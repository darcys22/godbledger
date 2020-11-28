package main

import (
	"context"
	"fmt"
	"time"

	"github.com/darcys22/godbledger/proto/transaction"

	"github.com/darcys22/godbledger/godbledger/cmd"
	"google.golang.org/grpc"

	"github.com/urfave/cli/v2"
)

var commandTagAccount = &cli.Command{
	Name:      "tag",
	Usage:     "ledger_cli tag <account> <tag>",
	ArgsUsage: "[]",
	Description: `
	Adds the tag specified in the second argument to the account specified in the first argument
`,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "delete",
			Aliases: []string{"d"},
			Usage:   "deletes tag rather than creates",
		},
	},
	Action: func(ctx *cli.Context) error {
		err, cfg := cmd.MakeConfig(ctx)
		if err != nil {
			return fmt.Errorf("Could not make config (%v)", err)
		}

		address := fmt.Sprintf("%s:%s", cfg.Host, cfg.RPCPort)
		log.WithField("address", address).Info("GRPC Dialing on port")
		opts := []grpc.DialOption{}

		if cfg.CACert != "" && cfg.Cert != "" && cfg.Key != "" {
			tlsCredentials, err := loadTLSCredentials(cfg)
			if err != nil {
				return fmt.Errorf("Could not load TLS credentials (%v)", err)
			}
			opts = append(opts, grpc.WithTransportCredentials(tlsCredentials))
		} else {
			opts = append(opts, grpc.WithInsecure())
		}

		// Set up a connection to the server.
		conn, err := grpc.Dial(address, opts...)
		if err != nil {
			return fmt.Errorf("Could not connect to GRPC (%v)", err)
		}
		defer conn.Close()
		client := transaction.NewTransactorClient(conn)

		ctxtimeout, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		if ctx.Bool("delete") {
			req := &transaction.DeleteTagRequest{
				Account: ctx.Args().Get(0),
				Tag:     ctx.Args().Get(1),
			}

			r, err := client.DeleteTag(ctxtimeout, req)
			if err != nil {
				return fmt.Errorf("Could not call Delete Tag Method (%v)", err)
			}

			log.Infof("Delete Tag Response: %s", r.GetMessage())
		} else {
			req := &transaction.TagRequest{
				Account: ctx.Args().Get(0),
				Tag:     ctx.Args().Get(1),
			}

			r, err := client.AddTag(ctxtimeout, req)
			if err != nil {
				return fmt.Errorf("Could not call Add Tag Method (%v)", err)
			}

			log.Infof("Create Tag Response: %s", r.GetMessage())
		}

		return nil
	},
}
