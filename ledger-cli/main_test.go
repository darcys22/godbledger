package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/darcys22/godbledger/godbledger/cmd"
	"github.com/darcys22/godbledger/tests/components"
	"github.com/darcys22/godbledger/tests/helpers"
	e2e "github.com/darcys22/godbledger/tests/params"

	"github.com/urfave/cli/v2"
)

func copyAndCapture(w io.Writer, r io.Reader) ([]byte, error) {
	var out []byte
	buf := make([]byte, 1024, 1024)
	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			d := buf[:n]
			out = append(out, d...)
			_, err := w.Write(d)
			if err != nil {
				return out, err
			}
		}
		if err != nil {
			// Read returns io.EOF at the end of file, which is not an error for us
			if err == io.EOF {
				err = nil
			}
			return out, err
		}
	}
}

func TestCommandLine(t *testing.T) {
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

	goDBLedgerPID := components.StartGoDBLedger(t, cfg, e2e.LogFileName, 0)
	defer helpers.KillProcesses(t, []int{goDBLedgerPID})

	time.Sleep(time.Duration(1) * time.Second)
	logfileName := fmt.Sprintf("%s-%d", e2e.LogFileName, 0)
	logFile, err := os.Open(logfileName)
	assert.NoError(t, err)
	defer helpers.DeleteLogFiles(t, []*os.File{logFile})

	if err := helpers.WaitForTextInFile(logFile, "Starting GoDBLedger Server"); err != nil {
		t.Fatalf("failed to find GoDBLedger start in logfile: %s, this means the server did not start: %v", logfileName, err)
	}

	// Start up ledger CLI and use the journal wizard
	cmd := exec.Command("ledger-cli", "journal")
	var stdout, stderr []byte
	var errStdout, errStderr error
	stdoutIn, _ := cmd.StdoutPipe()
	stdin, err := cmd.StdinPipe()
	assert.NoError(t, err)
	err = cmd.Start()
	assert.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)
	defer stdin.Close()
	go func() {
		io.WriteString(stdin, "2020-06-30\n")
		io.WriteString(stdin, "something\n")
		io.WriteString(stdin, "something\n")
		io.WriteString(stdin, "account1\n")
		io.WriteString(stdin, "300\n")
		io.WriteString(stdin, "y\n")
		io.WriteString(stdin, "something\n")
		io.WriteString(stdin, "account2\n")
		io.WriteString(stdin, "-300\n")
		io.WriteString(stdin, "n\n")
		stdout, errStdout = copyAndCapture(os.Stdout, stdoutIn)
		wg.Done()
	}()
	err = cmd.Wait()
	assert.NoError(t, err)
	wg.Wait()

	if errStdout != nil || errStderr != nil {
		t.Fatal("failed to capture stdout or stderr\n")
	}
	outStr, errStr := string(stdout), string(stderr)
	t.Logf("\nout:\n%s\nerr:\n%s\n", outStr, errStr)

	//TODO capture the success message
	//INFO ledger-cli: Add Transaction Response: c3jcc0m49b9ih5ol3e80

}
