## Go DB Ledger
An open source accounting system designed to have an easily accessable database and GRPC endpoints for communication.

## Executables

| Command         | Description                                                                                             |
|-----------------|---------------------------------------------------------------------------------------------------------|
| **`Godbledger`**    | The main server. It is the access point for transactions that will be saved to the accounting database. |
| `Ledgercli`     | A CLI client that can be used to transmit transactions to the server.                             |
| `Reporter`      | Builds basic reports from the database on the command line.                                             |


### Building the Proto Buffers
Call from the root directory
```
protoc -I proto/ proto/transaction.proto --go_out=plugins=grpc:proto
```

### SQL Querys the Database for the Transaction Listing
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
