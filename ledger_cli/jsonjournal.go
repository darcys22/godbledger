package main

import (
	"encoding/json"
	"fmt"

	"github.com/darcys22/godbledger/godbledger/cmd"

	"github.com/urfave/cli/v2"
)

var commandJSONJournal = &cli.Command{
	Name:      "jsonjournal",
	Usage:     "ledger_cli jsonjournal <journalInJSONFormat>",
	ArgsUsage: "[]",
	Description: `
	Creates a journal using the JSON passed through as the first Argument

	Example

	ledger_cli jsonjournal '{"Payee":"ijfjie","Date":"2019-06-30T00:00:00Z","AccountChanges":[{"Name":"Cash","Description":"jisfeij","Currency":"USD","Balance":"100"},{"Name":"Income","Description":"another","Currency":"USD","Balance":"-100"}],"Signature":"stuff"}'

`,
	Flags: []cli.Flag{},
	Action: func(ctx *cli.Context) error {
		err, cfg := cmd.MakeConfig(ctx)
		if err != nil {
			return fmt.Errorf("Could not make config (%v)", err)
		}

		// we initialize our request struct
		var req Transaction

		// we unmarshal our byteArray which contains our
		// jsonFile's content into 'users' which we defined above
		json.Unmarshal([]byte(ctx.Args().Get(0)), &req)

		log.Debugf("Transaction: %v\n", req)

		err = Send(cfg, &req)
		if err != nil {
			return fmt.Errorf("Could not send transaction (%v)", err)
		}

		return nil
	},
}
