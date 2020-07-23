// Package components defines utilities to spin up actual
// beacon node and validator processes as needed by end to end tests.
package components

import (
	"os/exec"
	//"strings"
	"testing"

	"github.com/darcys22/godbledger/godbledger/cmd"
)

func StartGoDBLedger(t *testing.T, config *cmd.LedgerConfig) int {

	args := []string{
		//fmt.Sprintf("--log-file=%s", stdOutFile.Name()),
		//"--verbosity=trace",
	}

	cmd := exec.Command("../build/bin/godbledger", args...)
	//t.Logf("Starting GoDBLedger %d with flags: %s", strings.Join(args[2:], " "))
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start GoDBLedger Server: %v", err)
	}

	//if err := helpers.WaitForTextInFile(stdOutFile, "RPC-API listening on port"); err != nil {
	//t.Fatalf("could not find multiaddr for node %d, this means the node had issues starting: %v", index, err)
	//}

	return cmd.Process.Pid
}
