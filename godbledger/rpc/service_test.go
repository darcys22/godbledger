package rpc

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/darcys22/godbledger/godbledger/cmd"
	"github.com/darcys22/godbledger/godbledger/ledger"

	"github.com/sirupsen/logrus"
	logTest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

type assertionLoggerFn func(string, ...interface{})

func init() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(ioutil.Discard)
}

func parseMsg(defaultMsg string, msg ...interface{}) string {
	if len(msg) >= 1 {
		msgFormat, ok := msg[0].(string)
		if !ok {
			return defaultMsg
		}
		return fmt.Sprintf(msgFormat, msg[1:]...)
	}
	return defaultMsg
}

// LogsContain checks whether a given substring is a part of logs. If flag=false, inverse is checked.
func LogsContain(loggerFn assertionLoggerFn, hook *logTest.Hook, want string, flag bool, msg ...interface{}) {
	_, file, line, _ := runtime.Caller(2)
	entries := hook.AllEntries()
	var logs []string
	match := false
	for _, e := range entries {
		msg, err := e.String()
		if err != nil {
			loggerFn("%s:%d Failed to format log entry to string: %v", filepath.Base(file), line, err)
			return
		}
		if strings.Contains(msg, want) {
			match = true
		}
		for _, field := range e.Data {
			fieldStr, ok := field.(string)
			if !ok {
				continue
			}
			if strings.Contains(fieldStr, want) {
				match = true
			}
		}
		logs = append(logs, msg)
	}
	var errMsg string
	if flag && !match {
		errMsg = parseMsg("Expected log not found", msg...)
	} else if !flag && match {
		errMsg = parseMsg("Unexpected log found", msg...)
	}
	if errMsg != "" {
		loggerFn("%s:%d %s: %v\nSearched logs:\n%v", filepath.Base(file), line, errMsg, want, logs)
	}
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

	LogsContain(t.Fatalf, hook, "GRPC Listening on port", true)
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

	LogsContain(t.Fatalf, hook, "GRPC Listening on port", true)
	LogsContain(t.Fatalf, hook, "You are using an insecure gRPC server", true)
	assert.NoError(t, rpcService.Stop())
}
