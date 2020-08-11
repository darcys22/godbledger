// Package cmd defines the command line flags for the shared utlities.
package cmd

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	"github.com/urfave/cli/v2"
)

func homeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}

// DefaultDataDir is the default data directory to use for the databases and other
// persistence requirements.
func DefaultDataDir() string {
	// Try to place the data folder in the user's home dir
	home := homeDir()
	if home != "" {
		if runtime.GOOS == "darwin" {
			return filepath.Join(home, "Library", "ledger")
		} else if runtime.GOOS == "windows" {
			return filepath.Join(home, "AppData", "Roaming", "ledger")
		} else {
			return filepath.Join(home, ".ledger")
		}
	}
	// As we cannot guess a stable location, return empty and handle later
	return ""
}

var (
	// VerbosityFlag defines the logrus configuration.
	VerbosityFlag = &cli.StringFlag{
		Name:  "verbosity",
		Usage: "Logging verbosity (debug, info=default, warn, error, fatal, panic)",
	}
	// DataDirFlag defines a path on disk.
	DataDirFlag = &cli.StringFlag{
		Name:  "datadir",
		Usage: "Data directory for the databases and keystore",
	}
	// ClearDB tells the node to remove any previously stored data at the data directory.
	ClearDB = &cli.BoolFlag{
		Name:  "clear-db",
		Usage: "Clears any previously stored data at the data directory",
	}
	ConfigFileFlag = &cli.StringFlag{
		Name:  "config",
		Usage: "TOML configuration file",
	}
	// RPCHost defines the host on which the RPC server should listen.
	RPCHost = &cli.StringFlag{
		Name:  "rpc-host",
		Usage: "Host on which the RPC server should listen",
	}
	// RPCPort defines a beacon node RPC port to open.
	RPCPort = &cli.StringFlag{
		Name:  "rpc-port",
		Usage: "RPC port exposed by GoDBLedger",
	}
	// CertFlag defines a flag for the node's TLS CA certificate.
	CACertFlag = &cli.StringFlag{
		Name:  "ca-cert",
		Usage: "Certificate Authority certificate for secure gRPC. Pass this and the tls-key flag in order to use gRPC securely.",
	}
	// CertFlag defines a flag for the node's TLS certificate.
	CertFlag = &cli.StringFlag{
		Name:  "tls-cert",
		Usage: "Certificate for secure gRPC. Pass this and the tls-key flag in order to use gRPC securely.",
	}
	// KeyFlag defines a flag for the node's TLS key.
	KeyFlag = &cli.StringFlag{
		Name:  "tls-key",
		Usage: "Key for secure gRPC. Pass this and the tls-cert flag in order to use gRPC securely.",
	}
)

func setConfig(ctx *cli.Context, cfg *LedgerConfig) {

	if ctx.IsSet(VerbosityFlag.Name) {
		cfg.LogVerbosity = ctx.String(VerbosityFlag.Name)
	}
	if ctx.IsSet(ConfigFileFlag.Name) {
		cfg.ConfigFile = ctx.String(ConfigFileFlag.Name)
	}
	if ctx.IsSet(DataDirFlag.Name) {
		cfg.ConfigFile = ctx.String(DataDirFlag.Name)
	}
	if ctx.IsSet(RPCHost.Name) {
		cfg.Host = ctx.String(RPCHost.Name)
	}
	if ctx.IsSet(RPCPort.Name) {
		cfg.RPCPort = ctx.String(RPCPort.Name)
	}
	if ctx.IsSet(CACertFlag.Name) {
		cfg.CACert = ctx.String(CACertFlag.Name)
	}
	if ctx.IsSet(CertFlag.Name) {
		cfg.Cert = ctx.String(CertFlag.Name)
	}
	if ctx.IsSet(KeyFlag.Name) {
		cfg.Key = ctx.String(KeyFlag.Name)
	}
}
