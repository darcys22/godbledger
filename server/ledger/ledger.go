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
	Name         []byte
	Users        []core.User
	Currencies   []core.Currency
	Chart        []core.Account
	Transactions []core.Transaction
	Splits       []core.Split

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
