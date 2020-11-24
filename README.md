# Go DB Ledger

[![Build Status]][Build Link] [![Book Status]][Book Link] [![Chat Badge]][Chat Link]

[Build Status]: https://travis-ci.com/darcys22/godbledger.svg?branch=dev
[Build Link]: https://github.com/darcys22/godbledger/actions
[Chat Badge]: https://img.shields.io/badge/chat-discord-%237289da
[Chat Link]: https://discord.gg/xHFufYC
[Book Status]:https://img.shields.io/badge/user--docs-master-informational
[Book Link]: https://github.com/darcys22/godbledger/wiki

GoDBLedger is an open source accounting system that aims to make the recording of double entry bookkeeping transactions programmable. It provide users with normal features that most finance systems tend to lack such as api endpoints for your scripts and a database backend with a clear schema so you can analyse your financial data using your software of choice. The ultimate goal is for your whole financial process to be automated from data entry to compilation of financials/tax returns.


#### How it works:
You are a business or individual wanting a system to record your profits and produce financial reports. You dont want to pay a cloud provider and you want to keep your financial data under your own control. You spin up a linux server (or raspberry pi) choose a database (Currently SQLite3 and MySQL are available) and you set up GoDBLedger to run on that server. You now have a place to send your double entry bookkeeping transactions which get saved into your own database! 

GoDBLedger gives you an api for the recording of transactions and there are  some command line binaries included to get you started.

[Watch the demo video](https://youtu.be/svyw9EOZuuc)

To get started view the quickstart on the wiki:
https://github.com/darcys22/godbledger/wiki/Quickstart


## Executables

| Command| Description |
| :----------------- | :---------------------------------------------------------------------------------------------------------| 
| **`Godbledger`**    | The main server. It is the access point for transactions that will be saved to the accounting database. |
| `Ledger_cli`     | A CLI client that can be used to transmit transactions to the server.                             |
| `Reporter`      | Builds basic reports from the database on the command line.                                             |

## Communicating with Godbledger and software examples

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

## Database and configuration

Godbledger will set a default configuration if none has been provided using Sqlite3 as the default database.

The config file can be found by default at:
```
~/.ledger/config.toml
```

## Building the Proto Buffers
Call from the root directory
```
protoc -I proto/ proto/transaction.proto --go_out=plugins=grpc:proto
```

## SQL Querys
default stored location for database is .ledger/ledgerdata `sqlite3 ledger.db`

**Select all transactions**
```
SELECT * FROM splits JOIN split_accounts ON splits.split_id = split_accounts.split_id

```

**Find the accounts with Tag**
```
SELECT * FROM accounts where account_id in (select account_id from account_tag where tag_id = 8);

```

## Contributing

### Local Development

1. Install golang version 1.13 or higher for your OS and architecture:

    - https://golang.org/doc/install

1. To build the `godbledger` executables natively for your OS and architecture you can simply use Make

    ```
    make
    ```

    The default make target is `build-native` which builds binaries native to your environment into the `./build/bin/native/` folder.

    NOTE: on windows you may need to install a C++ tool chain (e.g. [`tdm-gcc`](https://jmeubank.github.io/tdm-gcc/)) in order to cross compile the sqlite dependency.

    After building you can run the version you just built:

    ```
    ./build/bin/native/godbledger
    ```

1. Run the linter to discover any style or structural errors:

    ```
    make lint
    ```

1. Run the tests
   
    ```
    make test
    ```

    NOTE: the test suite depends on the `build-native` target as it includes an integration test which spins up an instance of `godbledger`

### Build architecture

The primary entrypoint into the build scripts is the `Makefile` which provides the aforementioned build targets:
- `build-native` (default)
- `lint`
- `test`

All three of which call into the `./utils/ci.go` script to do the actual work of setting up required env vars, building the executiables, and configuring output folders.

An additional `./utils/make-release.sh` script is available to help orchestrate the creation of zip/tarfiles.

### Cross-compiling with xgo/docker

In addition to the default, native build target, the `Makefile` also offers a `build-cross` target which uses a forked version of `xgo` (https://github.com/techknowlogick/xgo) to build for different operating systems and architectures, including linux variants, a MacOS-compatible binary, and windows-compatible exe files.

```
make build-cross
```

Go tooling natively offers cross-compiling features when the `CGO_ENABLED=0` flag is set; `godbledger`'s `go-sqlite3` dependency however requires `CGO_ENABLED=1` in order to link in the C-level bindings for SQLite.  Cross-compiling golang when `CGO` is enable is significantly more complicated as each platform and architecture can require a custom C++ toolchain.

`xgo` achieves consistency in cross-compilation using Docker, so running Docker Engine on your dev box is a requirement to running the `build-cross` target.

#### Install Docker Engine

The Docker web site includes detailed instructions on [installing and running Docker Engine](https://docs.docker.com/engine/install/) on a variety of supported platforms.

NOTE: if installing Docker Engine on a linux system make sure to follow the [Post-installation steps for Linux](https://docs.docker.com/engine/install/linux-postinstall/) in order to be able to run `docker` commands from local user accounts.

### TODO/Milestones
- ~~GoDBLedger server runs and accepts transactions~~
- ~~trial balance and transaction reports of journals~~
- ~~analyse database using metabase to create financial dashboard~~
- ~~authenticated api using mutual TLS~~
- web interface (GoDBLedger-Web)
- triple entry bookkeeping using signed transactions
- reconciliations and "bank feed"
- profit and loss and balance sheet reports
- Create Yurnell - programmable journal entries
