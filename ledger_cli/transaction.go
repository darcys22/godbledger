package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"math/big"
	"time"

	"github.com/darcys22/godbledger/godbledger/cmd"
	pb "github.com/darcys22/godbledger/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/urfave/cli/v2"
)

var commandSingleTestTransaction = &cli.Command{
	Name:      "single",
	Usage:     "submits a single transaction",
	ArgsUsage: "[]",
	Description: `
`,
	Flags: []cli.Flag{},
	Action: func(ctx *cli.Context) error {
		err, cfg := cmd.MakeConfig(ctx)
		if err != nil {
			log.Fatal(err)
		}

		date, _ := time.Parse("2006-01-02", "2011-03-15")
		desc := "Whole Food Market"

		transactionLines := make([]Account, 2)

		line1Account := "Expenses:Groceries"
		line1Desc := "Groceries"
		line1Amount := big.NewRat(7500, 1)

		transactionLines[0] = Account{
			Name:        line1Account,
			Description: line1Desc,
			Balance:     line1Amount,
			Currency:    "USD",
		}

		line2Account := "Assets:Checking"
		line2Desc := "Groceries"
		line2Amount := big.NewRat(-7500, 1)

		transactionLines[1] = Account{
			Name:        line2Account,
			Description: line2Desc,
			Balance:     line2Amount,
			Currency:    "USD",
		}

		req := &Transaction{
			Date:           date,
			Payee:          desc,
			AccountChanges: transactionLines,
			Signature:      "stuff",
		}

		err = Send(cfg, req)
		if err != nil {
			log.Fatalf("could not send: %v", err)
		}

		return nil
	},
}

func Send(cfg *cmd.LedgerConfig, t *Transaction) error {

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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	transactionLines := make([]*pb.LineItem, 2)

	for i, accChange := range t.AccountChanges {
		transactionLines[i] = &pb.LineItem{
			Accountname: accChange.Name,
			Description: accChange.Description,
			Amount:      accChange.Balance.Num().Int64(),
			Currency:    accChange.Currency,
		}
	}

	req := &pb.TransactionRequest{
		Date:        t.Date.Format("2006-01-02"),
		Description: t.Payee,
		Lines:       transactionLines,
		Signature:   t.Signature,
	}
	r, err := client.AddTransaction(ctx, req)
	if err != nil {
		log.Fatalf("Could not send transaction: %v", err)
	}
	log.Printf("Response: %s", r.GetMessage())
	return nil
}
func loadTLSCredentials(cfg *cmd.LedgerConfig) (credentials.TransportCredentials, error) {
	// Load certificate of the CA who signed server's certificate
	pemServerCA, err := ioutil.ReadFile(cfg.Cert)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}

	// Create the credentials and return it
	config := &tls.Config{
		RootCAs: certPool,
	}

	return credentials.NewTLS(config), nil
}
