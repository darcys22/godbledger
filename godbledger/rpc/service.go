package rpc

import (
	"context"
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/darcys22/godbledger/godbledger/ledger"
	pb "github.com/darcys22/godbledger/proto"
)

var log logrus.FieldLogger

func init() {
	log = logrus.WithField("prefix", "rpc")
}

type Service struct {
	ld         *ledger.Ledger
	ctx        context.Context
	cancel     context.CancelFunc
	grpcServer *grpc.Server
	listener   net.Listener
	port       string
	host       string
}

type Config struct {
	Port string
	Host string
}

func NewRPCService(ctx context.Context, cfg *Config, l *ledger.Ledger) *Service {
	ctx, cancel := context.WithCancel(ctx)
	return &Service{
		ld:     l,
		ctx:    ctx,
		cancel: cancel,
		port:   cfg.Port,
		host:   cfg.Host,
	}
}

// Start the gRPC server.
func (s *Service) Start() {
	log.Info("Starting service")
	address := fmt.Sprintf("%s:%s", s.host, s.port)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Errorf("Could not listen to port in Start() %s: %v", address, err)
	}
	s.listener = lis
	log.WithField("port", s.port).Info("Listening on port")
	s.grpcServer = grpc.NewServer()

	ledgerServer := &LedgerServer{ld: s.ld}

	pb.RegisterTransactorServer(s.grpcServer, ledgerServer)

	go func() {
		if s.listener != nil {
			if err := s.grpcServer.Serve(s.listener); err != nil {
				log.Errorf("Could not serve gRPC: %v", err)
			}
		}
	}()
}

// Stop the service.
func (s *Service) Stop() error {
	log.Info("Stopping service")
	s.cancel()
	if s.listener != nil {
		s.grpcServer.GracefulStop()
		log.Debug("Initiated graceful stop of gRPC server")
	}
	return nil
}

// Status returns nil or credentialError
func (s *Service) Status() error {
	return nil
}
