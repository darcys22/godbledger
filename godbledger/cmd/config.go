package cmd

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/urfave/cli/v2"

	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("prefix", "Config")

type LedgerConfig struct {
	Host             string // Host defines the address that the RPC will be opened on. Combined with RPC Port
	RPCPort          string // RPCPort defines the port that the server will listen for transactions on
	CACert           string // CACertFlag defines a flag for the server's Certificate Authority certificate (Public Key of Authority that signs clients Public Keys).
	Cert             string // CertFlag defines a flag for the server's TLS certificate (Servers Public Key to broadcast).
	Key              string // KeyFlag defines a flag for the server's TLS key (Servers Private Key).
	DataDirectory    string // DataDirectory defines the host systems folder directory holding the database and config files
	LogVerbosity     string // LogVerbosity defines the logging level {debug, info, warn, error, fatal, panic}
	ConfigFile       string // Location of the TOML config file, including directory path
	DatabaseType     string // Type of Database being used
	DatabaseLocation string // Location of the database file, including directory path or connection string
	PidFile          string // Location of the PID file, if blank will not be created
}

var (
	DumpConfigCommand = &cli.Command{
		Action:      dumpConfig,
		Name:        "dumpconfig",
		Usage:       "Show configuration values",
		ArgsUsage:   "",
		Category:    "MISCELLANEOUS COMMANDS",
		Description: `The dumpconfig command shows configuration values.`,
	}

	GenConfigCommand = &cli.Command{
		Action:      genConfig,
		Name:        "genconfig",
		Usage:       "godbledger genconfig [-m] [configFileLocation]",
		ArgsUsage:   "",
		Category:    "MISCELLANEOUS COMMANDS",
		Description: `The init command creates a configuration file at the optional [configFileLocation] else will default to HOME/.ledger/config.toml`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "mysql",
				Aliases: []string{"m"},
				Usage:   "set the database to use mysql rather than sqlite"},
		},
	}

	defaultLedgerConfig = &LedgerConfig{
		Host:             "127.0.0.1",
		RPCPort:          "50051",
		DataDirectory:    DefaultDataDir(),
		LogVerbosity:     "debug",
		ConfigFile:       DefaultDataDir() + "/config.toml",
		DatabaseType:     "sqlite3",
		DatabaseLocation: DefaultDataDir() + "/ledgerdata/ledger.db",
	}
)

func MakeConfig(cli *cli.Context) (error, *LedgerConfig) {
	config := defaultLedgerConfig
	//set logrus verbosity
	level, err := logrus.ParseLevel(config.LogVerbosity)
	if err != nil {
		log.Debugf("Error Parsing level: %s", err)
		return err, nil
	}

	logrus.SetLevel(level)
	setConfig(cli, config)
	if config.DataDirectory != DefaultDataDir() {
		config.ConfigFile = config.DataDirectory + "/config.toml"
		config.DatabaseLocation = config.DataDirectory + "/ledgerdata/ledger.db"
	} else {
		_, configerr := os.Stat("/var/lib/godbledger/config.toml")
		_, pidfileerr := os.Stat("/var/lib/godbledger/pidfile")
		if !os.IsNotExist(configerr) && !os.IsNotExist(pidfileerr) {
			config.DataDirectory = "/var/lib/godbledger"
			config.ConfigFile = config.DataDirectory + "/config.toml"
			config.DatabaseLocation = config.DataDirectory + "/ledgerdata/ledger.db"
		}
	}
	err = InitConfig(config)
	if err != nil {
		log.Debugf("Error initialising config: %s", err)
		return err, nil
	}

	if _, err := toml.DecodeFile(config.ConfigFile, &config); err != nil {
		log.Debugf("Error decoding config file: %s", err)
		return err, nil
	}

	//apply flags
	setConfig(cli, config)

	//set logrus verbosity
	level, err = logrus.ParseLevel(config.LogVerbosity)
	if err != nil {
		log.Debugf("Error Parsing level: %s", err)
		return err, nil
	}
	logrus.SetLevel(level)
	log.WithField("Config File", config.ConfigFile).Debug("Configuration Successfully loaded")

	return nil, config
}

func InitConfig(config *LedgerConfig) error {
	_, err := os.Stat(config.ConfigFile)
	if os.IsNotExist(err) {
		log.Debugf("Config File doesn't exist creating at %s", config.ConfigFile)
		os.MkdirAll(filepath.Dir(config.ConfigFile), os.ModePerm)
		buf := new(bytes.Buffer)
		if err := toml.NewEncoder(buf).Encode(config); err != nil {
			return err
		}
		dump, err := os.OpenFile(config.ConfigFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer dump.Close()
		dump.Write(buf.Bytes())
	}

	return nil
}

// dumpConfig is the dumpconfig command.
func dumpConfig(ctx *cli.Context) error {
	err, cfg := MakeConfig(ctx)
	if err != nil {
		log.Fatalf("Could not open the config file: %v", err)
		return err
	}

	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(cfg); err != nil {
		return err
	}

	dump := os.Stdout
	if ctx.NArg() > 0 {
		log.Infof("Writing Config to file: '%s'", ctx.Args().Get(0))
		dump, err = os.OpenFile(ctx.Args().Get(0), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Panicf("Could not open the file service: %v", err)
			return err
		}
		defer dump.Close()
	}
	dump.Write(buf.Bytes())

	return nil
}

// genConfig generates a default config file for godbledger.
func genConfig(ctx *cli.Context) error {
	config := defaultLedgerConfig
	if ctx.Bool("mysql") {
		config.DatabaseType = "mysql"
		config.DatabaseLocation = "godbledger:password@tcp(127.0.0.1:3306)/ledger?parseTime=true&charset=utf8"
	}

	if len(ctx.Args().Get(0)) > 0 {
		config.ConfigFile = ctx.Args().Get(0)
	}
	_, err := os.Stat(config.ConfigFile)
	if os.IsNotExist(err) {
		log.Infof("Config File doesn't exist creating at %s", config.ConfigFile)
		os.MkdirAll(filepath.Dir(config.ConfigFile), os.ModePerm)
		buf := new(bytes.Buffer)
		if err := toml.NewEncoder(buf).Encode(config); err != nil {
			return err
		}
		dump, err := os.OpenFile(config.ConfigFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer dump.Close()
		dump.Write(buf.Bytes())
	}

	return nil
}
