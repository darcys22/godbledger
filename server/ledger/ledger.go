package ledger

import (
	"path"

	"github.com/darcys22/godbledger/server/cmd"
	"github.com/darcys22/godbledger/server/core"
	"github.com/darcys22/godbledger/server/db"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const ledgerDBName = "ledgerdata"

var log = logrus.WithField("prefix", "ledger")

type Ledger struct {
	ledgerDb *db.LedgerDB
}

func New(ctx *cli.Context) (*Ledger, error) {
	baseDir := ctx.GlobalString(cmd.DataDirFlag.Name)
	log.Debug(cmd.DataDirFlag.Name)
	dbPath := path.Join(baseDir, ledgerDBName)
	log.WithField("path", dbPath).Info("Checking db path")
	if ctx.GlobalBool(cmd.ClearDB.Name) {
		if err := db.ClearDB(dbPath); err != nil {
			return nil, err
		}
	}

	ledgerDb, err := db.NewDB(dbPath)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	log.Info("Initialised database configuration")

	ledger := &Ledger{
		ledgerDb: ledgerDb,
	}

	return ledger, nil
}

func (l *Ledger) Insert(txn *core.Transaction) {
	log.Printf("Created Transaction: %s", txn)
	l.ledgerDb.SafeAddUser(txn.Poster)
	currencies, _ := l.GetCurrencies(txn)
	for _, currency := range currencies {
		l.ledgerDb.AddCurrency(currency)
	}
	accounts, _ := l.GetAccounts(txn)

	for _, account := range accounts {
		l.ledgerDb.AddAccount(account)
	}
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

func (l *Ledger) Start() {
	l.ledgerDb.InitDB()
}

func (l *Ledger) Stop() error {
	err := l.ledgerDb.Close()
	return err
}

func (l *Ledger) Status() error {
	return nil
}
