// Package endtoend performs full a end-to-end test for GoDBLedger,
// including spinning up a server and making sure its running
package tests

import (
	"context"
	//"path"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/darcys22/godbledger/godbledger/cmd"
	pb "github.com/darcys22/godbledger/proto"
	"github.com/darcys22/godbledger/tests/components"
	"github.com/darcys22/godbledger/tests/helpers"
	//e2e "github.com/darcys22/godbledger/tests/params"
	"google.golang.org/grpc"
)

func init() {
}

func runEndToEndTest(t *testing.T, config *cmd.LedgerConfig) {

	goDBLedgerPID := components.StartGoDBLedger(t, config)
	processIDs := []int{goDBLedgerPID}
	defer helpers.KillProcesses(t, processIDs)

	// Sleep depending on the count of validators, as generating the genesis state could take some time.
	time.Sleep(time.Duration(5) * time.Second)
	logFile, err := os.Open("logfile.txt")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Starting GoDBLedger ")
	t.Run("chain started", func(t *testing.T) {
		if err := helpers.WaitForTextInFile(logFile, "Starting GoDBLedger Server"); err != nil {
			t.Fatalf("failed to find GoDBLedger start in logs, this means the server did not start: %v", err)
		}
	})
	t.Log("Starting GoDBLedger ")

	//Failing early in case chain doesn't start.
	if t.Failed() {
		return
	}
	t.Log("Starting GoDBLedger before")

	conns := make([]*grpc.ClientConn, 1)
	for i := 0; i < len(conns); i++ {
		t.Logf("Starting GoDBLedger %d", i)
		conn, err := grpc.Dial(fmt.Sprintf("%s:%s", config.Host, config.RPCPort), grpc.WithInsecure())
		if err != nil {
			t.Fatalf("Failed to dial: %v", err)
		}
		t.Log("here before")
		conns[i] = conn
		t.Log("here after")
		defer func() {
			if err := conn.Close(); err != nil {
				t.Log(err)
			}
		}()
	}
	t.Log("Starting GoDBLedger after")

	client := pb.NewTransactorClient(conns[0])
	req := &pb.VersionRequest{
		Message: "Test",
	}
	_, err = client.NodeVersion(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	args := []string{
		"jsonjournal",
		`{"Payee":"ijfjie","Date":"2019-06-30T00:00:00Z","AccountChanges":[{"Name":"Cash","Description":"jisfeij","Currency":"USD","Balance":"100"},{"Name":"Income","Description":"another","Currency":"USD","Balance":"-100"}],"Signature":"stuff"}`,
	}

	cmd := exec.Command("../build/bin/ledger_cli", args...)

	//t.Logf("Sending jsonjournal with args %s", strings.Join(args[2:], " "))
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start ledger_cli: %v", err)
	}

}
