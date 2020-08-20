// Package endtoend performs full a end-to-end test for Prysm,
// including spinning up an ETH1 dev chain, sending deposits to the deposit
// contract, and making sure the beacon node and validators are running and
// performing properly for a few epochs.
package tests

import (
	//"fmt"
	//"os"
	"os/exec"
	//"path"
	"testing"
	//"time"

	"github.com/darcys22/godbledger/godbledger/cmd"
	"github.com/darcys22/godbledger/tests/components"
	"github.com/darcys22/godbledger/tests/helpers"
	//"google.golang.org/grpc"
)

func init() {
}
func runEndToEndTest(t *testing.T, config *cmd.LedgerConfig) {

	goDBLedgerPID := components.StartGoDBLedger(t, config)
	processIDs := []int{goDBLedgerPID}
	defer helpers.KillProcesses(t, processIDs)

	// Sleep depending on the count of validators, as generating the genesis state could take some time.
	time.Sleep(time.Duration(5) * time.Second)
	logFile, err := os.Open(logfile.txt)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("chain started", func(t *testing.T) {
		if err := helpers.WaitForTextInFile(logFile, "Starting GoDBLedger Server"); err != nil {
			t.Fatalf("failed to find GoDBLedger start in logs, this means the server did not start: %v", err)
		}
	})

	//Failing early in case chain doesn't start.
	if t.Failed() {
		return
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
