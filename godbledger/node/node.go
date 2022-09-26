package node

import (
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/darcys22/godbledger/godbledger/cmd"
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
	PidFile  string
}

func New(ctx *cli.Context, cfg *cmd.LedgerConfig) (*Node, error) {
	registry := core.NewServiceRegistry()

	ledger := &Node{
		ctx:      ctx,
		services: registry,
		stop:     make(chan struct{}),
		PidFile:  cfg.PidFile,
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
		"version": version.VersionWithCommit(),
	}).Info("Starting GoDBLedger Server")

	n.writePIDFile()
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
	if len(n.PidFile) > 0 {
		_ = os.Remove(n.PidFile)
	}
	n.services.StopAll()
	close(n.stop)
}

// writePIDFile retrieves the current process ID and writes it to file.
func (n *Node) writePIDFile() {
	if n.PidFile == "" {
		return
	}

	// Ensure the required directory structure exists.
	err := os.MkdirAll(filepath.Dir(n.PidFile), 0700)
	if err != nil {
		log.Error("Failed to verify pid directory", "error", err)
		os.Exit(1)
	}

	// Retrieve the PID and write it to file.
	pid := strconv.Itoa(os.Getpid())
	if err := os.WriteFile(n.PidFile, []byte(pid), 0644); err != nil {
		log.Error("Failed to write pidfile", "error", err)
		os.Exit(1)
	}

	log.Info("Writing PID file", "path", n.PidFile, "pid", pid)
}
