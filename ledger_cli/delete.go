package main

import (
	"context"
	"fmt"
	"time"

	"github.com/darcys22/godbledger/godbledger/cmd"

	pb "github.com/darcys22/godbledger/proto"
	"google.golang.org/grpc"

	//"github.com/urfave/cli"
	"github.com/urfave/cli/v2"
)

var commandDeleteTransaction = &cli.Command{
	Name:      "delete",
	Usage:     "ledger_cli delete <transaction_id>",
	ArgsUsage: "[]",
	Description: `
	Deletes a transaction from the database
`,
	Flags: []cli.Flag{},
	Action: func(ctx *cli.Context) error {
		err, cfg := cmd.MakeConfig(ctx)
		if err != nil {
			log.Fatal(err)
		}

		if ctx.NArg() > 0 {
			address := fmt.Sprintf("%s:%s", cfg.Host, cfg.RPCPort)
			log.WithField("address", address).Info("GRPC Dialing on port")
			opts := []grpc.DialOption{}

			if cfg.CACert != "" && cfg.Cert != "" && cfg.Key != "" {
				tlsCredentials, err := loadTLSCredentials(cfg)
				if err != nil {
					log.Fatal("cannot load TLS credentials: ", err)
				}
				opts = append(opts, grpc.WithTransportCredentials(tlsCredentials))
			} else {
				opts = append(opts, grpc.WithInsecure())
			}

			// Set up a connection to the server.
			conn, err := grpc.Dial(address, opts...)
			if err != nil {
				log.Fatalf("did not connect: %v", err)
			}
			defer conn.Close()
			client := pb.NewTransactorClient(conn)

			ctxtimeout, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			signature := "test"
			req := &pb.DeleteRequest{
				Identifier: ctx.Args().Get(0),
				Signature:  signature,
			}
			r, err := client.DeleteTransaction(ctxtimeout, req)
			if err != nil {
				log.Fatalf("could not delete: %v", err)
			}
			log.Printf("Response: %s", r.GetMessage())
			return nil
		} else {
			log.Fatalf("This command requires an argument.")
		}

		return nil
	},
}

var commandVoidTransaction = &cli.Command{
	Name:      "void",
	Usage:     "ledger_cli void <transaction_id>",
	ArgsUsage: "[]",
	Description: `
	Reverses a transaction by creating an identical inverse and tags both transactions as void 
`,
	Flags: []cli.Flag{},
	Action: func(ctx *cli.Context) error {
		err, cfg := cmd.MakeConfig(ctx)
		if err != nil {
			log.Fatal(err)
		}

		if ctx.NArg() > 0 {
			address := fmt.Sprintf("%s:%s", cfg.Host, cfg.RPCPort)
			log.WithField("address", address).Info("GRPC Dialing on port")
			opts := []grpc.DialOption{}

			if cfg.CACert != "" && cfg.Cert != "" && cfg.Key != "" {
				tlsCredentials, err := loadTLSCredentials(cfg)
				if err != nil {
					log.Fatal("cannot load TLS credentials: ", err)
				}
				opts = append(opts, grpc.WithTransportCredentials(tlsCredentials))
			} else {
				opts = append(opts, grpc.WithInsecure())
			}

			// Set up a connection to the server.
			conn, err := grpc.Dial(address, opts...)
			if err != nil {
				log.Fatalf("did not connect: %v", err)
			}
			defer conn.Close()
			client := pb.NewTransactorClient(conn)

			ctxtimeout, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			signature := "test"
			req := &pb.DeleteRequest{
				Identifier: ctx.Args().Get(0),
				Signature:  signature,
			}
			r, err := client.VoidTransaction(ctxtimeout, req)
			if err != nil {
				log.Fatalf("could not void: %v", err)
			}
			log.Printf("Response: %s", r.GetMessage())
			return nil
		} else {
			log.Fatalf("This command requires an argument.")
		}

		return nil
	},
}
