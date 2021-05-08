// secure connection test establishes a connection using mutual tls, and sending test transaction to verify

// +build integration,secure

package tests

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/darcys22/godbledger/proto/transaction"

	"github.com/darcys22/godbledger/godbledger/cmd"
	"github.com/darcys22/godbledger/tests/components"

	"github.com/darcys22/godbledger/tests/helpers"

	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func TestSecureConnection(t *testing.T) {
	// Create a config from the defaults which would usually be created by the CLI library
	set := flag.NewFlagSet("test", 0)
	set.String("config", "", "doc")
	ctx := cli.NewContext(nil, set, nil)
	err, cfg := cmd.MakeConfig(ctx)
	if err != nil {
		t.Fatalf("New Config Failed: %v", err)
	}

	// Set the Database type to a SQLite3 in memory database
	cfg.DatabaseType = "memorydb"

	// Set the RPC port to random higher port to not clash with other tests
	cfg.RPCPort = "55051"

	// The Certificates and Private Keys necessary for both the server and the client has previously been generated using utils/gen.sh these have been saved in the certs directory in this test folder.

	// Add the servers credential filenames to the configuration
	cfg.CACert = "certs/ca-cert.pem"
	cfg.Cert = "certs/server-cert.pem"
	cfg.Key = "certs/server-key.pem"

	// Also add the clients credentials to variables for later usage
	clientCertFilename := "certs/client-cert.pem"
	clientKeyFilename := "certs/client-key.pem"

	processIDs := []int{}
	logFiles := []*os.File{}
	goDBLedgerPID := components.StartGoDBLedger(t, cfg, "secure-connection.log", 0)
	processIDs = append(processIDs, goDBLedgerPID)
	time.Sleep(time.Duration(1) * time.Second)
	logfileName := fmt.Sprintf("%s-%d", "secure-connection.log", 0)
	logFile, err := os.Open(logfileName)
	if err != nil {
		t.Fatal(err)
	}
	logFiles = append(logFiles, logFile)

	t.Run("Server Started", func(t *testing.T) {
		if err := helpers.WaitForTextInFile(logFile, "Starting GoDBLedger Server"); err != nil {
			t.Fatalf("failed to find GoDBLedger start in logfile: %s, this means the server did not start: %v", logfileName, err)
		}
	})

	//Failing early in case chain doesn't start.
	if t.Failed() {
		return
	}
	defer helpers.KillProcesses(t, processIDs)
	defer helpers.DeleteLogFiles(t, logFiles)

	t.Logf("Starting GoDBLedger")
	port, _ := strconv.Atoi(cfg.RPCPort)

	opts := []grpc.DialOption{}

	// Load certificate of the CA who signed server's certificate
	pemServerCA, err := ioutil.ReadFile(cfg.CACert)
	if err != nil {
		t.Fatalf("Failed reading CA certificate: %v", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		t.Fatal("failed to add CA's certificate to pool")
	}

	// Load client's certificate and private key
	clientCert, err := tls.LoadX509KeyPair(clientCertFilename, clientKeyFilename)
	if err != nil {
		t.Fatalf("Failed reading Client certificate and key: %v", err)
	}

	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
	}

	tlsCredentials := credentials.NewTLS(config)
	opts = append(opts, grpc.WithTransportCredentials(tlsCredentials))

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", cfg.Host, port), opts...)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			t.Log(err)
		}
	}()

	client := transaction.NewTransactorClient(conn)
	req := &transaction.VersionRequest{
		Message: "Test",
	}
	_, err = client.NodeVersion(context.Background(), req)
	if err != nil {
		t.Fatalf("Node Version request failed: %v", err)
	}

}
