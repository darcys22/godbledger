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
	//time.Sleep(time.Duration(params.BeaconConfig().GenesisDelay) * time.Second)
	//beaconLogFile, err := os.Open(path.Join(e2e.TestParams.LogPath, fmt.Sprintf(e2e.BeaconNodeLogFileName, 0)))
	//if err != nil {
	//t.Fatal(err)
	//}

	//t.Run("chain started", func(t *testing.T) {
	//if err := helpers.WaitForTextInFile(beaconLogFile, "Chain started within the last epoch"); err != nil {
	//t.Fatalf("failed to find chain start in logs, this means the chain did not start: %v", err)
	//}
	//})

	// Failing early in case chain doesn't start.
	//if t.Failed() {
	//return
	//}

	args := []string{
		"jsonjournal",
		`{"Payee":"ijfjie","Date":"2019-06-30T00:00:00Z","AccountChanges":[{"Name":"Cash","Description":"jisfeij","Currency":"USD","Balance":"100"},{"Name":"Income","Description":"another","Currency":"USD","Balance":"-100"}],"Signature":"stuff"}`,
	}

	cmd := exec.Command("../build/bin/ledger_cli", args...)

	//t.Logf("Sending jsonjournal with args %s", strings.Join(args[2:], " "))
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start ledger_cli: %v", err)
	}

	//bt := new(testMatcher)

	//bt.walk(t, basicTestDir, func(t *testing.T, name string, test *BasicTest) {
	//if err := bt.checkFailure(t, name, test.Run(false)); err != nil {
	//t.Errorf("test failed: %v", err)
	//}
	//})
}
