## Go DB Ledger
An open source accounting system designed to have an easily accessable database and GRPC endpoints for communication.

## Executables

| Command         | Description |
| :----------------- | ---------------------------------------------------------------------------------------------------------:
| | **`Godbledger`**    | The main server. It is the access point for transactions that will be saved to the accounting database. |
| `Ledgercli`     | A CLI client that can be used to transmit transactions to the server.                             |
| `Reporter`      | Builds basic reports from the database on the command line.                                             |

### Communicating with Godbledger and software examples

**GRPC and Proto Buffers**
The primary way to communicate with Godbledger is through the GRPC endpoint, submitting a transaction that contains your journal entry/transaction.

**Ledgercli** communicates with Godbledger using GRPC but opens up several cli commands for usage

**Ledger files** the ledgercli allows for the processing of [ledger files](https://www.ledger-cli.org/). This has been roughly implemented by forking https://github.com/howeyc/ledger

**Trading Simulator** 
An [example project](https://github.com/darcys22/Trading-Simulator) has been developed that simulates a market trader (Random Walk) and the trades are recorded using Godbledger

**Reporter**
The general usage of Godbledger is not to provide information but to simply guide transactions to be recorded in a consistent manner in the database. To actually view your financial information we should query the database directly. Reporter has two SQL queries in built (Transaction Listing, and Trial Balance) that will be formatted in a table/json/csv for your viewing.

**PDF Financial Statements**
Reporter also has a function to generate pdf financial reports. Two templates for Profit and Loss and Balance sheet have been provided.

The PDF files are generated from [handlebars](https://handlebarsjs.com/) iterating over the tagged accounts. This is compiled into PDF using nodejs.

templates can be viewed [here](https://github.com/darcys22/pdf-generator)

### Database and configuration

Godbledger will set a default configuration if none has been provided using Sqlite3 as the default database.

The config file can be found by default at:
```
~/.ledger/config.toml
```

### Building the Proto Buffers
Call from the root directory
```
protoc -I proto/ proto/transaction.proto --go_out=plugins=grpc:proto
```

### SQL Querys
default stored location for database is .ledger/ledgerdata `sqlite3 ledger.db`

**Select all transactions**
```
SELECT * FROM splits JOIN split_accounts ON splits.split_id = split_accounts.split_id

```

**Find the accounts with Tag**
```
SELECT * FROM accounts where account_id in (select account_id from account_tag where tag_id = 8);

```

### TODO
- Add a call that returns the trial balance with all the tags on each account
- Create another server that monitors the system and updates programmable entries
- Add an edit transaction function
