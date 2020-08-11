package main

import (
	"context"
	"fmt"
	"time"

	"github.com/darcys22/godbledger/godbledger/cmd"
	pb "github.com/darcys22/godbledger/proto"

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
			log.Fatal(err)
		}

		address := fmt.Sprintf("%s:%s", cfg.Host, cfg.RPCPort)
		log.WithField("address", address).Info("GRPC Dialing on port")
		opts := []grpc.DialOption{}

		if cfg.Cert != "" && cfg.Key != "" {
			tlsCredentials, err := loadTLSCredentials(cfg)
			if err != nil {
				log.Fatal("cannot load TLS credentials: ", err)
			}
			opts = append(opts, grpc.WithTransportCredentials(tlsCredentials))
		} else {
			opts = append(opts, grpc.WithInsecure())
		}

		// Set up a connection to the server.
		//conn, err := grpc.Dial(address, grpc.WithInsecure())
		conn, err := grpc.Dial(address, opts...)
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
		client := pb.NewTransactorClient(conn)

		ctxtimeout, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		if ctx.Bool("delete") {
			req := &pb.DeleteTagRequest{
				Account:   ctx.Args().Get(0),
				Tag:       ctx.Args().Get(1),
				Signature: "blah",
			}

			r, err := client.DeleteTag(ctxtimeout, req)
			if err != nil {
				log.Fatalf("could not greet: %v", err)
			}

			log.Printf("Delete Tag Response: %s", r.GetMessage())
		} else {
			req := &pb.TagRequest{
				Account:   ctx.Args().Get(0),
				Tag:       ctx.Args().Get(1),
				Signature: "blah",
			}

			r, err := client.AddTag(ctxtimeout, req)
			if err != nil {
				log.Fatalf("could not greet: %v", err)
			}

			log.Printf("Create Tag Response: %s", r.GetMessage())
		}

		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}

		return nil
	},
}
