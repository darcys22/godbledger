// Package endtoend performs full a end-to-end test for GoDBLedger,
// including spinning up a server and making sure its running, and sending test data to verify
package tests

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	pb "github.com/darcys22/godbledger/proto"

	"github.com/darcys22/godbledger/godbledger/cmd"
	"github.com/darcys22/godbledger/tests/components"

	ev "github.com/darcys22/godbledger/tests/evaluators"
	"github.com/darcys22/godbledger/tests/helpers"
	e2e "github.com/darcys22/godbledger/tests/params"
	"github.com/darcys22/godbledger/tests/types"

	"google.golang.org/grpc"
)

func runEndToEndTest(t *testing.T, config *cmd.LedgerConfig) {

	goDBLedgerPID := components.StartGoDBLedger(t, config)
	processIDs := []int{goDBLedgerPID}
	defer helpers.KillProcesses(t, processIDs)

	time.Sleep(time.Duration(1) * time.Second)
	logFile, err := os.Open(e2e.LogFileName)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Server Started", func(t *testing.T) {
		if err := helpers.WaitForTextInFile(logFile, "Starting GoDBLedger Server"); err != nil {
			t.Fatalf("failed to find GoDBLedger start in logs, this means the server did not start: %v", err)
		}
	})

	//Failing early in case chain doesn't start.
	if t.Failed() {
		return
	}

	conns := make([]*grpc.ClientConn, 1)
	for i := 0; i < len(conns); i++ {
		t.Logf("Starting GoDBLedger %d", i)
		conn, err := grpc.Dial(fmt.Sprintf("%s:%s", config.Host, config.RPCPort), grpc.WithInsecure())
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

	client := pb.NewTransactorClient(conns[0])
	req := &pb.VersionRequest{
		Message: "Test",
	}
	_, err = client.NodeVersion(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	evaluators := []types.Evaluator{ev.SingleTransaction}
	for _, evaluator := range evaluators {
		t.Run(evaluator.Name, func(t *testing.T) {
			if err := evaluator.Evaluation(conns...); err != nil {
				t.Errorf("evaluation failed for sync node: %v", err)
			}
		})
	}

}
