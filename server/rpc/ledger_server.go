package rpc

import (
	"context"

	"github.com/darcys22/godbledger/server/version"

	pb "github.com/darcys22/godbledger/proto"
)

type LedgerServer struct{}

func (s *LedgerServer) AddTransaction(ctx context.Context, in *pb.TransactionRequest) (*pb.TransactionResponse, error) {
	log.Printf("Received New Transaction Request")
	return &pb.TransactionResponse{Message: "Accepted"}, nil
}

func (s *LedgerServer) NodeVersion(ctx context.Context, in *pb.VersionRequest) (*pb.VersionResponse, error) {
	log.Printf("Received Version Request: %s", in)
	return &pb.VersionResponse{Message: version.Version}, nil
}
