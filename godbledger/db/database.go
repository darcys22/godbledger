package db

import (
	"database/sql"
	"time"

	"github.com/darcys22/godbledger/godbledger/core"
)

// Database wraps all database operations.
type Database interface {
	InitDB() error
	Close() error
	AddTransaction(txn *core.Transaction) (string, error)
	FindTransaction(txnID string) (*core.Transaction, error)
	DeleteTransaction(txnID string) error
	FindTag(tag string) (int, error)
	AddTag(tag string) error
	SafeAddTag(tag string) error
	SafeAddTagToAccount(account, tag string) error
	AddTagToAccount(accountID string, tag int) error
	DeleteTagFromAccount(account, tag string) error
	SafeAddTagToTransaction(txnID, tag string) error
	AddTagToTransaction(txnID string, tag int) error
	DeleteTagFromTransaction(txnID, tag string) error
	FindCurrency(cur string) (*core.Currency, error)
	AddCurrency(cur *core.Currency) error
	SafeAddCurrency(cur *core.Currency) error
	DeleteCurrency(currency string) error
	FindAccount(code string) (*core.Account, error)
	AddAccount(*core.Account) error
	SafeAddAccount(*core.Account) error
	FindUser(pubKey string) (*core.User, error)
	AddUser(usr *core.User) error
	SafeAddUser(usr *core.User) error
	GetTB(date time.Time) (*[]core.TBAccount, error)
	GetListing(startdate, enddate time.Time) (*[]core.Transaction, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
}
