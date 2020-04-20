## Go DB Ledger
An open source accounting system designed to have an easily accessable database and GRPC endpoints for communication.

Checkout the [Wiki](https://github.com/darcys22/godbledger/wiki)

[Quickstart](https://github.com/darcys22/godbledger/wiki/Quickstart)

## Executables

| Command| Description |
| :----------------- | :---------------------------------------------------------------------------------------------------------| 
| **`Godbledger`**    | The main server. It is the access point for transactions that will be saved to the accounting database. |
| `Ledger_cli`     | A CLI client that can be used to transmit transactions to the server.                             |
| `Reporter`      | Builds basic reports from the database on the command line.                                             |

### Communicating with Godbledger and software examples

**GRPC and Proto Buffers**
The primary way to communicate with Godbledger is through the GRPC endpoint, submitting a transaction that contains your journal entry/transaction.

a python client with example calls can be found [here](https://github.com/darcys22/godbledger-pythonclient)

**Ledger_cli** included with this repo communicates with Godbledger using GRPC and gives some convenient CLI commands

**Ledger files** `ledger_cli` allows for the processing of [ledger files](https://www.ledger-cli.org/). This has been roughly implemented by forking https://github.com/howeyc/ledger

**Trading Simulator** 
An [example project](https://github.com/darcys22/Trading-Simulator) has been developed that simulates a market trader bot and the trades are recorded using Godbledger

**Reporter**
The general usage of Godbledger is not to provide information but to simply guide transactions to be recorded in a consistent manner in the database. To actually view your financial information we should query the database directly. Reporter has two SQL queries in built (Transaction Listing, and Trial Balance) that will be formatted in a table/json/csv for your viewing.

```
reporter trialbalance
reporter transactions
```

**PDF Financial Statements**
Reporter also has a function to generate pdf financial reports. Two templates for a Profit and Loss and a Balance sheet have been provided.

```
reporter pdf -template profitandloss
```

The PDF files are generated from [handlebars](https://handlebarsjs.com/) iterating over the tagged accounts. This is compiled into PDF using nodejs.

Templates can be viewed [here](https://github.com/darcys22/pdf-generator)

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
- Create Yurnell - programmable journal entries
- run GoDBLedger on a separate server and access the open port through the network
