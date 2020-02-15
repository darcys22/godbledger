package cmd

import (
	"bytes"
	//"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/urfave/cli"

	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("prefix", "Config")

type LedgerConfig struct {
	RPCPort          string // RPCPort defines the port that the server will listen for transactions on
	DataDirectory    string // DataDirectory defines the host systems folder directory holding the database and config files
	LogVerbosity     string // LogVerbosity defines the logging level {debug, info, warn, error, fatal, panic}
	ConfigFile       string // Location of the TOML config file, including directory path
	DatabaseType     string // Type of Database being used
	DatabaseLocation string // Location of the database file, including directory path
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

	defaultLedgerConfig = &LedgerConfig{
		RPCPort:          "50051",
		DataDirectory:    DefaultDataDir(),
		LogVerbosity:     "debug",
		ConfigFile:       DefaultDataDir() + "/config.toml",
		DatabaseType:     "sqlite3",
		DatabaseLocation: DefaultDataDir() + "/ledgerdata/ledger.db",
	}
)

func MakeConfig(cli *cli.Context) (error, *LedgerConfig) {

	log.Infof("Setting up configuration")
	config := defaultLedgerConfig
	//set logrus verbosity
	level, err := logrus.ParseLevel(config.LogVerbosity)
	if err != nil {
		log.Debugf("Error Parsing level: %s", err)
		return err, nil
	}
	logrus.SetLevel(level)
	if len(cli.String("config")) > 0 {
		config.ConfigFile = cli.String("config")
	}

	log.Debugf("Filepath to config file: %s", config.ConfigFile)
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

	return nil, config
}

func InitConfig(config *LedgerConfig) error {
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

// dumpConfig is the dumpconfig command.
func dumpConfig(ctx *cli.Context) error {

	err, cfg := MakeConfig(ctx)

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
