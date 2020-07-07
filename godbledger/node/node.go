package node

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"
	//"github.com/urfave/cli"
	"github.com/urfave/cli/v2"

	"github.com/darcys22/godbledger/godbledger/core"
	"github.com/darcys22/godbledger/godbledger/db"
	"github.com/darcys22/godbledger/godbledger/version"
)

var log = logrus.WithField("prefix", "node")

type Node struct {
	ctx      *cli.Context
	lock     sync.RWMutex
	services *core.ServiceRegistry
	stop     chan struct{} // Channel to wait for termination notifications.
	DB       *db.Database
}

func New(ctx *cli.Context) (*Node, error) {

	registry := core.NewServiceRegistry()

	ledger := &Node{
		ctx:      ctx,
		services: registry,
		stop:     make(chan struct{}),
	}

	return ledger, nil

}

func (n *Node) Register(constructor core.Service) error {
	n.lock.Lock()
	defer n.lock.Unlock()

	n.services.RegisterService(constructor)

	return nil
}

func (n *Node) Start() {
	n.lock.Lock()
	log.WithFields(logrus.Fields{
		"version": version.Version,
	}).Info("Starting ledger node")

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
	close(n.stop)
}
