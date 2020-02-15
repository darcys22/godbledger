package db

import (
	"github.com/darcys22/godbledger/godbledger/core"
)

// Database wraps all database operations.
type Database interface {
	InitDB() error
	Close() error
	AddTransaction(txn *core.Transaction) error
	DeleteTransaction(txnID string) error
	FindTag(tag string) (int, error)
	AddTag(tag string) error
	SafeAddTag(tag string) error
	SafeAddTagToAccount(account, tag string) error
	AddTagToAccount(accountID string, tag int) error
	DeleteTagFromAccount(account, tag string) error
	FindCurrency(cur string) (*core.Currency, error)
	AddCurrency(cur *core.Currency) error
	SafeAddCurrency(cur *core.Currency) error
	FindAccount(code string) (*core.Account, error)
	AddAccount(*core.Account) error
	SafeAddAccount(*core.Account) error
	FindUser(pubKey string) (*core.User, error)
	AddUser(usr *core.User) error
	SafeAddUser(usr *core.User) error
}
