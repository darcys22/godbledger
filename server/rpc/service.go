package rpc

import "fmt"

var log logrus.FieldLogger

func init() {
	log = logrus.WithField("prefix", "rpc")
}

//type Service struct {
//ctx context.Context
//cancel context.CancelFunc
//port	string
//listener	net.Listener

//grpcServer *grpc.Server

//}

//type Config struct {
//Port string
//}

//type testServer struct {
//savedFeatures []*pb.Feature // read-only after initialized

//mu         sync.Mutex // protects routeNotes
//routeNotes map[string][]*pb.RouteNote
//}

//func NewRPCService(ctx context.Context, cfg *Config) *Service {
//ctx, cancel := context.WithCancel(ctx)
//return &Service {
//ctx: ctx,
//cancel: cancel,
//port: cfg.Port,
//}
//}

//func (s *Service) Start() {
//log.Info("Starting service")
//lis, err := net.Listen("tcp", fmt.Sprintf(":%s", s.port))
//if err != nil {
//log.Errorf("Could not listen to port in Start() :%s", s.port, err)
//}
//s.listener = lis
//log.WithField("port", s.port).Info("Listening on port")

//var opts []grpc.ServerOption
//grpcServer := grpc.NewServer(opts...)
//grpcServer.RegisterService(,testServer)

//}
//}

//func main() {
//fmt.Println("vim-go")
//}
