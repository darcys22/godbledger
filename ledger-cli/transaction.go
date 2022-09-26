package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/darcys22/godbledger/proto/transaction"

	"github.com/darcys22/godbledger/godbledger/cmd"
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
			return fmt.Errorf("Could not make config (%v)", err)
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
		}

		err = Send(cfg, req)
		if err != nil {
			return fmt.Errorf("Could not send transaction (%v)", err)
		}

		return nil
	},
}

func Send(cfg *cmd.LedgerConfig, t *Transaction) error {
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
	//conn, err := grpc.Dial(address, grpc.WithInsecure())
	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		return fmt.Errorf("Could not connect to GRPC (%v)", err)
	}
	defer conn.Close()
	client := transaction.NewTransactorClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	transactionLines := make([]*transaction.LineItem, 2)

	for i, accChange := range t.AccountChanges {
		amountInt64 := accChange.Balance.Num().Int64() * int64(100) / accChange.Balance.Denom().Int64()
		transactionLines[i] = &transaction.LineItem{
			Accountname: accChange.Name,
			Description: accChange.Description,
			Amount:      amountInt64,
			Currency:    accChange.Currency,
		}
	}

	req := &transaction.TransactionRequest{
		Date:        t.Date.Format("2006-01-02"),
		Description: t.Payee,
		Lines:       transactionLines,
	}
	r, err := client.AddTransaction(ctx, req)
	if err != nil {
		return fmt.Errorf("Could not call Add Transaction Method (%v)", err)
	}
	log.Infof("Add Transaction Response: %s", r.GetMessage())
	return nil
}
func loadTLSCredentials(cfg *cmd.LedgerConfig) (credentials.TransportCredentials, error) {
	// Load certificate of the CA who signed server's certificate
	pemServerCA, err := os.ReadFile(cfg.CACert)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}

	// Load client's certificate and private key
	clientCert, err := tls.LoadX509KeyPair(cfg.Cert, cfg.Key)
	if err != nil {
		return nil, err
	}

	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
	}

	return credentials.NewTLS(config), nil
}
