## Installation

### Building
```
```

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
