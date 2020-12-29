// secure connection test establishes a connection using mutual tls, and sending test transaction to verify

// +build integration

package tests

import (
	"context"
	"flag"
	"fmt"
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
)

func secure_connection_test(t *testing.T) {

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

	processIDs := []int{}
	logFiles := []*os.File{}
	goDBLedgerPID := components.StartGoDBLedger(t, cfg, 1)
	processIDs = append(processIDs, goDBLedgerPID)
	time.Sleep(time.Duration(1) * time.Second)
	logfileName := fmt.Sprintf("%s-%d", "secure-connection.log", 1)
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

	opts := []grpc.ServerOption{}
	if s.withCACert != "" && s.withCert != "" && s.withKey != "" {
		// Load certificate of the CA who signed client's certificate
		pemClientCA, err := ioutil.ReadFile(s.withCACert)
		if err != nil {
			return nil, err
		}
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(pemClientCA) {
			return nil, fmt.Errorf("failed to add client CA's certificate")
		}

		// Load server's certificate and private key
		serverCert, err := tls.LoadX509KeyPair(s.withCert, s.withKey)
		if err != nil {
			return nil, err
		}

		// Create the credentials and return it
		config := &tls.Config{
			Certificates: []tls.Certificate{serverCert},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    certPool,
		}

		creds := credentials.NewTLS(config)
		opts = append(opts, grpc.Creds(creds))
	} else {
		t.Fatal("You are using an insecure gRPC server. This test was designed to test a secure connection")
	}

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
		t.Fatal(err)
	}
}
