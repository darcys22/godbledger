package node

import (
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"godbledger/cmd"
	"godbledger/core"
	"godbledger/db"
	"godbledger/version"
)

var log = logrus.WithField("prefix", "node")

const ledgerDBName = "ledgerdata"

type Node struct {
	//ledger *ledger.LedgerDB
	ctx      *cli.Context
	lock     sync.RWMutex
	services *core.ServiceRegistry
	stop     chan struct{} // Channel to wait for termination notifications.
	DB       *db.LedgerDB
}

func New(ctx *cli.Context) (*Node, error) {
	//l, err := NewLedgerDB()
	//if err != nil {
	//return nil, err
	//}

	registry := core.NewServiceRegistry()

	ledger := &Node{
		//ledger: l,
		ctx:      ctx,
		services: registry,
		stop:     make(chan struct{}),
	}

	return ledger, nil

}

func (n *Node) Start() {
	n.lock.Lock()
	log.WithFields(logrus.Fields{
		"version": version.Version,
	}).Info("Starting ledger node")

	ldb, _ := db.NewDB("")
	n.DB = ldb

	n.DB.TestDB()

	n.services.StartAll()
	stop := n.stop
	n.lock.Unlock()

	go func() {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(sigc)
		<-sigc
		log.Info("Got interrupt, shutting down...")
		go n.Close()
		for i := 10; i > 0; i-- {
			<-sigc
			if i > 1 {
				log.Info("Already shutting down, interrupt more to panic", "times", i-1)
			}
		}
		panic("Panic closing the node")
	}()

	<-stop
}

// Close handles graceful shutdown of the system.
func (n *Node) Close() {
	n.lock.Lock()
	defer n.lock.Unlock()

	log.Info("Stopping ledger node")
	n.services.StopAll()
	//b.services.StopAll()
	//if err := b.db.Close(); err != nil {
	//log.Errorf("Failed to close database: %v", err)
	//}
	close(n.stop)
}

func (n *Node) startDB(ctx *cli.Context) error {
	baseDir := ctx.GlobalString(cmd.DataDirFlag.Name)
	dbPath := path.Join(baseDir, ledgerDBName)
	if n.ctx.GlobalBool(cmd.ClearDB.Name) {
		if err := db.ClearDB(dbPath); err != nil {
			return err
		}
	}

	db, err := db.NewDB(dbPath)
	if err != nil {
		return err
	}

	log.WithField("path", dbPath).Info("Checking db")
	n.DB = db
	return nil
}
