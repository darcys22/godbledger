package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/urfave/cli"
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
	Action: func(c *cli.Context) error {

		// we initialize our request struct
		var req Transaction

		// we unmarshal our byteArray which contains our
		// jsonFile's content into 'users' which we defined above
		json.Unmarshal([]byte(c.Args().Get(0)), &req)

		fmt.Printf("%v\n", req)

		err := Send(&req)
		if err != nil {
			log.Fatalf("could not send: %v", err)
		}

		return nil
	},
}
