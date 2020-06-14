package ledger

import (
	"path"
	"strings"
	"time"

	"github.com/darcys22/godbledger/godbledger/cmd"
	"github.com/darcys22/godbledger/godbledger/core"
	"github.com/darcys22/godbledger/godbledger/db"
	"github.com/darcys22/godbledger/godbledger/db/mysql"
	"github.com/darcys22/godbledger/godbledger/db/sqlite3"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
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
	case "sqlite3":

		log.Debug("Using Sqlite3")
		dbPath := path.Join(cfg.DataDirectory, ledgerDBName)
		log.WithField("path", dbPath).Debug("Checking db path")
		if ctx.Bool(cmd.ClearDB.Name) {
			if err := sqlite3db.ClearDB(dbPath); err != nil {
				return nil, err
			}
		}
		ledgerdb, err := sqlite3db.NewDB(dbPath)
		ledger.LedgerDb = ledgerdb
		if err != nil {
			return nil, err
		}
	case "mysql":
		log.Debug("Using MySQL")
		ledgerdb, err := mysqldb.NewDB(cfg.DatabaseLocation)
		//if ctx.Bool(cmd.ClearDB.Name) {
		//if err := ledgerdb.ClearDB(ledgerDBName); err != nil {
		//return nil, err
		//}
		//}
		ledger.LedgerDb = ledgerdb
		if err != nil {
			return nil, err
		}
	case "memorydb":
		log.Info("Using in memory database")
		log.Fatal("In memory database not implemented")
	default:
		log.Println("No implementation available for that database.")
	}

	log.Debug("Initialised database configuration")

	return ledger, nil
}

func (l *Ledger) Insert(txn *core.Transaction) (string, error) {
	log.Info("Created Transaction: %s", txn)
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
	log.Infof("Deleting Transaction: %s", txnID)
	l.LedgerDb.DeleteTransaction(txnID)
}

func (l *Ledger) Void(txnID string) error {
	log.Infof("Voiding Transaction: %s", txnID)
	return l.LedgerDb.SafeAddTagToTransaction(txnID, "Void")
}

func (l *Ledger) InsertTag(account, tag string) error {
	log.Infof("Creating Tag %s on %s", tag, account)
	return l.LedgerDb.SafeAddTagToAccount(account, tag)
}

func (l *Ledger) DeleteTag(account, tag string) error {
	log.Infof("Deleting Tag %s from %s", tag, account)
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

		if exists == false {
			currencies = append(currencies, cur)
		}

	}

	return currencies, nil
}

func (l *Ledger) InsertCurrency(curr *core.Currency) error {
	log.Infof("Creating Currency %s with %s decimals", curr.Name, curr.Decimals)
	return l.LedgerDb.SafeAddCurrency(curr)
}

func (l *Ledger) DeleteCurrency(currency string) error {
	log.Infof("Deleting Currency %s", currency)
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
			if exists == false {
				accounts = append(accounts, a)
			}
		}

	}

	return accounts, nil
}

func (l *Ledger) GetTB(date time.Time) (*[]core.TBAccount, error) {
	return l.LedgerDb.GetTB(date)
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
