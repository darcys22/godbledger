package main

import (
	"context"
	"log"
	"time"

	pb "github.com/darcys22/godbledger/proto"
	"google.golang.org/grpc"

	"github.com/urfave/cli"
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
			Name:  "delete, d",
			Usage: "deletes tag rather than creates",
		},
	},
	Action: func(c *cli.Context) error {

		// Set up a connection to the server.
		conn, err := grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
		client := pb.NewTransactorClient(conn)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		if c.Bool("delete") {
			req := &pb.DeleteTagRequest{
				Account:   c.Args().Get(0),
				Tag:       c.Args().Get(1),
				Signature: "blah",
			}

			r, err := client.DeleteTag(ctx, req)
			if err != nil {
				log.Fatalf("could not greet: %v", err)
			}

			log.Printf("Delete Tag Response: %s", r.GetMessage())
		} else {
			req := &pb.TagRequest{
				Account:   c.Args().Get(0),
				Tag:       c.Args().Get(1),
				Signature: "blah",
			}

			r, err := client.AddTag(ctx, req)
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
