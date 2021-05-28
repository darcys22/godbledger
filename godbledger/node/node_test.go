package node

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/darcys22/godbledger/godbledger/cmd"
	"github.com/darcys22/godbledger/godbledger/ledger"
	"github.com/darcys22/godbledger/internal"

	"github.com/sirupsen/logrus"
	logTest "github.com/sirupsen/logrus/hooks/test"
	"github.com/urfave/cli/v2"
)

const (
	maxPollingWaitTime = 1 * time.Second
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(ioutil.Discard)
}

// Test that godbledger node can close.
func TestNodeClose_OK(t *testing.T) {
	hook := logTest.NewGlobal()

	app := cli.App{}
	set := flag.NewFlagSet("test", 0)
	set.String("config", "", "doc")

	ctx := cli.NewContext(&app, set, nil)

	node, err := New(ctx)
	assert.NoError(t, err)

	node.Close()

	internal.LogsContain(t.Fatalf, hook, "Stopping ledger node", true)
}

// TestClearDB tests clearing the database
func TestClearDB(t *testing.T) {
	hook := logTest.NewGlobal()

	randPath, err := rand.Int(rand.Reader, big.NewInt(1000000))
	assert.NoError(t, err, "Could not generate random number for file path")
	tmp := filepath.Join(os.TempDir(), fmt.Sprintf("datadirtest%d", randPath))
	assert.NoError(t, os.RemoveAll(tmp))

	app := cli.App{}
	set := flag.NewFlagSet("test", 0)
	set.Bool(cmd.ClearDB.Name, true, "")

	ctx := cli.NewContext(&app, set, nil)
	assert.NoError(t, err)
	err, cfg := cmd.MakeConfig(ctx)
	assert.NoError(t, err)
	cfg.DatabaseType = "memorydb"
	cfg.DataDirectory = tmp

	node, err := New(ctx)
	assert.NoError(t, err)

	ledger, err := ledger.New(ctx, cfg)
	assert.NoError(t, err)

	node.Register(ledger)
	go node.Start()
	d := time.Now().Add(maxPollingWaitTime)
	contextWithDeadline, cancel := context.WithDeadline(context.Background(), d)
	defer cancel()
	<-contextWithDeadline.Done()
	internal.LogsContain(t.Fatalf, hook, "Clearing SQLite3 DB", true)
	//case <-contextWithDeadline.Done():

	node.Close()
	assert.NoError(t, os.RemoveAll(tmp))
}
