package node

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"godbledger/db"
	"godbledger/ledger"
	"godbledger/version"
)

var log = logrus.WithField("prefix", "node")

type Node struct {
	//ledger *ledger.LedgerDB
	ctx  *cli.Context
	lock sync.RWMutex
	stop chan struct{} // Channel to wait for termination notifications.
	db   *db.LedgerDB
}

func New(ctx *cli.Context) (*Node, error) {
	//l, err := NewLedgerDB()
	//if err != nil {
	//return nil, err
	//}

	ledger := &Node{
		//ledger: l,
		ctx:  ctx,
		stop: make(chan struct{}),
	}

	return &Node{}

}

func (n *Node) Start() error {
	n.lock.Lock()
	log.WithFields(logrus.Fields{
		"version": version.Version(),
	}).Info("Starting ledger node")

	stop := l.stop
	n.lock.Unlock()

	go func() {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(sigc)
		<-sigc
		log.Info("Got interrupt, shutting down...")
		debug.Exit(b.ctx) // Ensure trace and CPU profile data are flushed.
		go b.Close()
		for i := 10; i > 0; i-- {
			<-sigc
			if i > 1 {
				log.Info("Already shutting down, interrupt more to panic", "times", i-1)
			}
		}
		panic("Panic closing the beacon node")
	}()

	<-stop
}

// Close handles graceful shutdown of the system.
func (n *Node) Close() {
	n.lock.Lock()
	defer n.lock.Unlock()

	log.Info("Stopping ledger node")
	//b.services.StopAll()
	//if err := b.db.Close(); err != nil {
	//log.Errorf("Failed to close database: %v", err)
	//}
	close(n.stop)
}

func (n *Node) startDB(ctx *cli.Context) error {
	baseDir := ctx.GlobalString(cmd.DataDirFlag.Name)
	dbPath := path.Join(baseDir, beaconChainDBName)
	if b.ctx.GlobalBool(cmd.ClearDB.Name) {
		if err := db.ClearDB(dbPath); err != nil {
			return err
		}
	}

	db, err := db.NewDB(dbPath)
	if err != nil {
		return err
	}

	log.WithField("path", dbPath).Info("Checking db")
	b.db = db
	return nil
}
