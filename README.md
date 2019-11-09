## Go DB Ledger
Accounting system designed to have an easily accessable database

## Executables

| Command         | Description                                                                                             |
|-----------------|---------------------------------------------------------------------------------------------------------|
| **`server`**    | The main server. It is the access point for transactions that will be saved to the accounting database. |
| `ledgercli`     | Gives a CLI client that can be used to transmit transactions to the server.                             |
| `reporter`      | Builds basic reports from the database on the command line.                                             |


### Building the Proto Buffers
Call from the root directory
```
protoc -I proto/ proto/transaction.proto --go_out=plugins=grpc:proto
```

### Query the Database for the Transaction Listing
default stored location for database is .ledger/ledgerdata
```
sqlite3 ledger.db
SELECT * FROM splits JOIN split_accounts ON splits.split_id = split_accounts.split_id

```
