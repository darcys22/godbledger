package rpc

import (
	"context"
	"errors"
	"flag"
	"io/ioutil"
	"testing"

	"github.com/darcys22/godbledger/godbledger/cmd"
	"github.com/darcys22/godbledger/godbledger/ledger"
	"github.com/darcys22/godbledger/internal"

	"github.com/sirupsen/logrus"
	logTest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(ioutil.Discard)
}

func TestLifecycle_OK(t *testing.T) {
	hook := logTest.NewGlobal()

	set := flag.NewFlagSet("test", 0)
	set.String("config", "", "doc")
	ctx := cli.NewContext(nil, set, nil)
	err, cfg := cmd.MakeConfig(ctx)
	assert.NoError(t, err)

	cfg.DatabaseType = "memorydb"
	cfg.Host = "127.0.0.1"
	cfg.RPCPort = "7348"
	cfg.CACert = "bob.crt"
	cfg.Cert = "alice.crt"
	cfg.Key = "alice.key"

	ledger, err := ledger.New(ctx, cfg)
	assert.NoError(t, err)

	rpcService := NewRPCService(context.Background(), &Config{
		Host:       cfg.Host,
		Port:       cfg.RPCPort,
		CACertFlag: cfg.CACert,
		CertFlag:   cfg.Cert,
		KeyFlag:    cfg.Key,
	}, ledger)

	rpcService.Start()

	internal.LogsContain(t.Fatalf, hook, "GRPC Listening on port", true)
	assert.NoError(t, rpcService.Stop())
}

func TestStatus_CredentialError(t *testing.T) {
	credentialErr := errors.New("credentialError")
	s := &Service{credentialError: credentialErr}

	assert.Contains(t, s.credentialError.Error(), s.Status().Error())
}

func TestRPC_InsecureEndpoint(t *testing.T) {
	hook := logTest.NewGlobal()

	set := flag.NewFlagSet("test", 0)
	set.String("config", "", "doc")
	ctx := cli.NewContext(nil, set, nil)
	err, cfg := cmd.MakeConfig(ctx)
	assert.NoError(t, err)

	cfg.DatabaseType = "memorydb"
	cfg.Host = "127.0.0.1"
	cfg.RPCPort = "7777"

	ledger, err := ledger.New(ctx, cfg)
	assert.NoError(t, err)

	rpcService := NewRPCService(context.Background(), &Config{
		Host: cfg.Host,
		Port: cfg.RPCPort,
	}, ledger)

	rpcService.Start()

	internal.LogsContain(t.Fatalf, hook, "GRPC Listening on port", true)
	internal.LogsContain(t.Fatalf, hook, "You are using an insecure gRPC server", true)
	assert.NoError(t, rpcService.Stop())
}
