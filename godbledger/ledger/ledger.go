package ledger

import (
	"path"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/darcys22/godbledger/godbledger/cmd"
	"github.com/darcys22/godbledger/godbledger/core"
	"github.com/darcys22/godbledger/godbledger/db"
	"github.com/darcys22/godbledger/godbledger/db/mysqldb"
	"github.com/darcys22/godbledger/godbledger/db/sqlite3db"
)

const ledgerDBName = "ledgerdata"

var log = logrus.WithField("prefix", "ledger")

type Ledger struct {
	LedgerDb db.Database
	Config   *cmd.LedgerConfig
}

func New(ctx *cli.Context, cfg *cmd.LedgerConfig) (*Ledger, error) {

	ledger := &Ledger{
		Config: cfg,
	}

	switch strings.ToLower(cfg.DatabaseType) {
	case "sqlite3", "memorydb":

		log.Debug("Using Sqlite3")
		mode := "rwc"
		dbPath := path.Join(cfg.DataDirectory, ledgerDBName)
		if strings.ToLower(cfg.DatabaseType) == "memorydb" {
			log.Debug("In Memory only Mode")
			mode = "memory"
		}
		log.WithField("path", dbPath).Debug("Checking db path")
		if ctx.Bool(cmd.ClearDB.Name) {
			log.Info("Clearing SQLite3 DB")
			if err := sqlite3db.ClearDB(dbPath); err != nil {
				return nil, err
			}
		}
		ledgerdb, err := sqlite3db.NewDB(dbPath, mode)
		ledger.LedgerDb = ledgerdb
		if err != nil {
			return nil, err
		}
	case "mysql":
		log.Debug("Using MySQL")
		ledgerdb, err := mysqldb.NewDB(cfg.DatabaseLocation)
		if ctx.Bool(cmd.ClearDB.Name) {
			log.Info("Clearing MySQL DB")
			if err := ledgerdb.ClearDB(); err != nil {
				return nil, err
			}
		}
		ledger.LedgerDb = ledgerdb
		if err != nil {
			return nil, err
		}
	default:
		log.Fatal("No implementation available for that database.")
	}

	log.Debug("Initialised database configuration")

	return ledger, nil
}

func (l *Ledger) Insert(txn *core.Transaction) (string, error) {
	log.WithField("transaction", txn).Debug("Created Transaction")
	l.LedgerDb.SafeAddUser(txn.Poster)
	currencies, _ := l.GetCurrencies(txn)
	for _, currency := range currencies {
		l.LedgerDb.SafeAddCurrency(currency)
	}
	accounts, _ := l.GetAccounts(txn)

	for _, account := range accounts {
		l.LedgerDb.SafeAddAccount(account)
		l.LedgerDb.SafeAddTagToAccount(account.Name, "main")
	}

	response, err := l.LedgerDb.AddTransaction(txn)
	if err != nil {
		return "", err
	}

	return response, nil
}

func (l *Ledger) Delete(txnID string) {
	l.LedgerDb.DeleteTransaction(txnID)
}

func (l *Ledger) Void(txnID string, usr *core.User) error {
	txn, err := l.LedgerDb.FindTransaction(txnID)
	if err != nil {
		return err
	}

	log.Debugf("Transaction Found to Void: %+v", txn)

	newTxn, err := core.ReverseTransaction(txn, usr)
	if err != nil {
		return err
	}

	log.Debugf("Reversed Transaction: %+v", newTxn)

	newJournalID, err := l.Insert(newTxn)
	if err != nil {
		return err
	}
	log.Debug("Successful insert of reversing transaction")

	err = l.LedgerDb.SafeAddTagToTransaction(newJournalID, "Void")
	if err != nil {
		return err
	}
	log.Debug("New Transaction Tagged Void")

	err = l.LedgerDb.SafeAddTagToTransaction(txnID, "Void")
	if err != nil {
		return err
	}
	log.Debug("Original Transaction Tagged Void")

	return nil
}

func (l *Ledger) InsertTag(account, tag string) error {
	return l.LedgerDb.SafeAddTagToAccount(account, tag)
}

func (l *Ledger) DeleteTag(account, tag string) error {
	return l.LedgerDb.DeleteTagFromAccount(account, tag)
}

func (l *Ledger) GetCurrencies(txn *core.Transaction) ([]*core.Currency, error) {

	currencies := []*core.Currency{}

	for _, split := range txn.Splits {
		cur := split.Currency
		exists := false

		for _, b := range currencies {
			if b == cur {
				exists = true
			}
		}

		if !exists {
			currencies = append(currencies, cur)
		}

	}

	return currencies, nil
}

func (l *Ledger) InsertCurrency(curr *core.Currency) error {
	return l.LedgerDb.SafeAddCurrency(curr)
}

func (l *Ledger) DeleteCurrency(currency string) error {
	return l.LedgerDb.DeleteCurrency(currency)
}

func (l *Ledger) GetAccounts(txn *core.Transaction) ([]*core.Account, error) {
	accounts := []*core.Account{}

	for _, split := range txn.Splits {
		accs := split.Accounts

		for _, a := range accs {
			exists := false
			for _, b := range accounts {
				if b == a {
					exists = true
				}
			}
			if !exists {
				accounts = append(accounts, a)
			}
		}

	}

	return accounts, nil
}

func (l *Ledger) GetTB(date time.Time) (*[]core.TBAccount, error) {
	return l.LedgerDb.GetTB(date)
}

func (l *Ledger) GetListing(enddate, startdate time.Time) (*[]core.Transaction, error) {
	return l.LedgerDb.GetListing(enddate, startdate)
}

func (l *Ledger) Start() {
	l.LedgerDb.InitDB()
}

func (l *Ledger) Stop() error {
	err := l.LedgerDb.Close()
	return err
}

func (l *Ledger) Status() error {
	return nil
}
