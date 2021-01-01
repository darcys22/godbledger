// Package endtoend performs full a end-to-end test for GoDBLedger,
// including spinning up a server and making sure its running, and sending test data to verify

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

	ev "github.com/darcys22/godbledger/tests/evaluators"
	"github.com/darcys22/godbledger/tests/helpers"
	e2e "github.com/darcys22/godbledger/tests/params"
	"github.com/darcys22/godbledger/tests/types"

	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
)

// This slice will contain all the evaluators (tests) that we want to run. These are defined in the /evaluators subdirectory.
var evaluators = []types.Evaluator{
	ev.SingleTransaction,
	ev.DoubleTransaction,
	ev.DoubleTransactionIgnoreOne,
}

func TestEndToEnd_MinimalConfig(t *testing.T) {

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

	processIDs := []int{}
	logFiles := []*os.File{}
	for i := 0; i < len(evaluators); i++ {
		goDBLedgerPID := components.StartGoDBLedger(t, cfg, e2e.LogFileName, i)
		processIDs = append(processIDs, goDBLedgerPID)
		time.Sleep(time.Duration(1) * time.Second)
		logfileName := fmt.Sprintf("%s-%d", e2e.LogFileName, i)
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
	}
	defer helpers.KillProcesses(t, processIDs)
	defer helpers.DeleteLogFiles(t, logFiles)

	conns := make([]*grpc.ClientConn, len(evaluators))
	for i := 0; i < len(conns); i++ {
		t.Logf("Starting GoDBLedger %d", i)
		port, _ := strconv.Atoi(cfg.RPCPort)
		conn, err := grpc.Dial(fmt.Sprintf("%s:%d", cfg.Host, port+i), grpc.WithInsecure())
		if err != nil {
			t.Fatalf("Failed to dial: %v", err)
		}
		conns[i] = conn
		defer func() {
			if err := conn.Close(); err != nil {
				t.Log(err)
			}
		}()
	}

	client := transaction.NewTransactorClient(conns[0])
	req := &transaction.VersionRequest{
		Message: "Test",
	}
	_, err = client.NodeVersion(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	for i, evaluator := range evaluators {
		t.Run(evaluator.Name, func(t *testing.T) {
			if err := evaluator.Evaluation(conns[i]); err != nil {
				t.Errorf("evaluation failed for sync node: %v", err)
			}
		})
	}

}
