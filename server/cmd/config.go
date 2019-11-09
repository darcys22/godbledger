package cmd

import (
	"bytes"
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/urfave/cli"
)

type LedgerConfig struct {
	RPCPort       string // RPCPort defines the port that the server will listen for transactions on
	DataDirectory string // DataDirectory defines the host systems folder directory holding the database and config files
	LogVerbosity  string // LogVerbosity defines the logging level {debug, info, warn, error, fatal, panic}
	ConfigFile    string // Location of the TOML config file, including directory path
}

var defaultLedgerConfig = &LedgerConfig{
	RPCPort:       "50051",
	DataDirectory: DefaultDataDir(),
	LogVerbosity:  "debug",
	ConfigFile:    DefaultDataDir() + "/config.toml",
}

func makeConfig(cli *cli.Context) (error, *LedgerConfig) {

	config := defaultLedgerConfig
	if _, err := toml.DecodeFile("example.toml", &config); err != nil {
		return err, nil
	}

	return nil, config
}

func InitConfig() error {
	config := defaultLedgerConfig
	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(config); err != nil {
		return err
	}
	fmt.Println(buf.String())
	return nil
}
