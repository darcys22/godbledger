// Package components defines utilities to spin up actual
// beacon node and validator processes as needed by end to end tests.
package components

import (
	"os/exec"

	//"strings"
	"fmt"
	"strings"
	"testing"

	"github.com/darcys22/godbledger/godbledger/cmd"
	"github.com/darcys22/godbledger/tests/helpers"
	e2e "github.com/darcys22/godbledger/tests/params"
)

func StartGoDBLedger(t *testing.T, config *cmd.LedgerConfig) int {
	stdOutFile, err := helpers.DeleteAndCreateFile("", e2e.LogFileName)

	if err != nil {
		t.Fatal(err)
	}

	args := []string{
		fmt.Sprintf("--log-file=%s", stdOutFile.Name()),
		"--verbosity=trace",
		fmt.Sprintf("--rpc-host=%s", config.Host),
		fmt.Sprintf("--rpc-port=%s", config.RPCPort),
		fmt.Sprintf("--database=%s", config.DatabaseType),
		fmt.Sprintf("--database-location=%s", config.DatabaseLocation),
	}

	cmd := exec.Command("../build/bin/native/godbledger", args...)
	t.Logf("Starting GoDBLedger with flags: %s", strings.Join(args[:], " "))
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start GoDBLedger Server: %v", err)
	}

	if err := helpers.WaitForTextInFile(stdOutFile, "GRPC Listening on port"); err != nil {
		t.Fatalf("could not find GRPC starting for server, this means the server had issues starting: %v", err)
	}

	return cmd.Process.Pid
}
