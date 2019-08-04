package ledger

import (
	"godbledger/core"
	"godbledger/db"

	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("prefix", "ledger")

type Ledger struct {
	Name         []byte
	Users        []core.User
	Currencies   []core.Currency
	Chart        []core.Account
	Transactions []core.Transaction
	Splits       []core.Split

	ledgerDb *db.LedgerDB
}

func New() (*Ledger, error) {
	ledgerDb, err := db.NewDB("")
	if err != nil {
		return nil, err
	}
	log.Info("Initialised database configuration")

	ledger := &Ledger{
		ledgerDb: ledgerDb,
	}

	return ledger, nil
}

func (l *Ledger) Start() {
	l.ledgerDb.TestDB()
}

func (l *Ledger) Stop() error {
	err := l.ledgerDb.Close()
	return err
}

func (l *Ledger) Status() error {
	return nil
}

func (l *Ledger) AppendTransaction(txn *core.Transaction) error {
	l.Transactions = append(l.Transactions, *txn)
	return nil
}

func (l *Ledger) AppendUser(usr core.User) error {
	l.Users = append(l.Users, usr)
	return nil
}
