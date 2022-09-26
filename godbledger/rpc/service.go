package rpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"

	"github.com/darcys22/godbledger/proto/transaction"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"

	"github.com/darcys22/godbledger/godbledger/ledger"
)

var log logrus.FieldLogger

func init() {
	log = logrus.WithField("prefix", "rpc")
}

type Service struct {
	ld              *ledger.Ledger
	ctx             context.Context
	cancel          context.CancelFunc
	grpcServer      *grpc.Server
	listener        net.Listener
	port            string
	host            string
	withCACert      string
	withCert        string
	withKey         string
	credentialError error
}

type Config struct {
	Port       string
	Host       string
	CACertFlag string
	CertFlag   string
	KeyFlag    string
}

func NewRPCService(ctx context.Context, cfg *Config, l *ledger.Ledger) *Service {
	ctx, cancel := context.WithCancel(ctx)
	return &Service{
		ld:         l,
		ctx:        ctx,
		cancel:     cancel,
		port:       cfg.Port,
		host:       cfg.Host,
		withCACert: cfg.CACertFlag,
		withCert:   cfg.CertFlag,
		withKey:    cfg.KeyFlag,
	}
}

// Start the gRPC server.
func (s *Service) Start() {
	log.Debug("Starting service")
	address := fmt.Sprintf("%s:%s", s.host, s.port)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Errorf("Could not listen to port in Start() %s: %v", address, err)
	}
	s.listener = lis
	log.WithField("address", address).Info("GRPC Listening on port")

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(s.unaryConnectionInterceptor),
		grpc.StreamInterceptor(s.streamConnectionInterceptor),
	}

	if s.withCACert != "" && s.withCert != "" && s.withKey != "" {
		creds, err := s.loadTLSCredentials()
		if err != nil {
			log.Errorf("Could not load TLS keys: %s", err)
			s.credentialError = err
		}
		opts = append(opts, grpc.Creds(creds))
	} else {
		log.Warn("You are using an insecure gRPC server. If you are running your GoDBLedger Server and " +
			"client on the same machine, you can ignore this message.")
	}

	s.grpcServer = grpc.NewServer(opts...)

	ledgerServer := &LedgerServer{ld: s.ld}

	transaction.RegisterTransactorServer(s.grpcServer, ledgerServer)

	// Register reflection service on gRPC server.
	reflection.Register(s.grpcServer)

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
	if s.credentialError != nil {
		return s.credentialError
	}
	return nil
}

// Stream interceptor for new stream client connections to GRPC.
func (s *Service) streamConnectionInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	_ *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	s.logNewClientConnection(ss.Context())
	return handler(srv, ss)
}

// Unary interceptor for new unary client connections to GRPC.
func (s *Service) unaryConnectionInterceptor(
	ctx context.Context,
	req interface{},
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	s.logNewClientConnection(ctx)
	return handler(ctx, req)
}

func (s *Service) logNewClientConnection(ctx context.Context) {
	if clientInfo, ok := peer.FromContext(ctx); ok {
		log.WithFields(logrus.Fields{
			"addr": clientInfo.Addr.String(),
		}).Infof("New gRPC client connected to GoDBLedger")
	}
}

func (s *Service) loadTLSCredentials() (credentials.TransportCredentials, error) {
	// Load certificate of the CA who signed client's certificate
	pemClientCA, err := os.ReadFile(s.withCACert)
	if err != nil {
		return nil, err
	}
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemClientCA) {
		return nil, fmt.Errorf("failed to add client CA's certificate")
	}

	// Load server's certificate and private key
	serverCert, err := tls.LoadX509KeyPair(s.withCert, s.withKey)
	if err != nil {
		return nil, err
	}

	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}

	return credentials.NewTLS(config), nil
}
