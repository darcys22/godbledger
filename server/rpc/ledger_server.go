package rpc

import (
	"context"
	"math/big"
	"time"

	"github.com/darcys22/godbledger/server/core"
	"github.com/darcys22/godbledger/server/ledger"
	"github.com/darcys22/godbledger/server/version"

	pb "github.com/darcys22/godbledger/proto"
)

type LedgerServer struct {
	ld *ledger.Ledger
}

func (s *LedgerServer) AddTransaction(ctx context.Context, in *pb.TransactionRequest) (*pb.TransactionResponse, error) {
	log.Printf("Received New Transaction Request")

	usr, err := core.NewUser("Sean")
	if err != nil {
		log.Error(err)
	}
	aud, err := core.NewCurrency("AUD", 2)
	if err != nil {
		log.Error(err)
	}

	txn, err := core.NewTransaction(usr)
	if err != nil {
		log.Error(err)
	}
	txn.Description = []byte(in.GetDescription())

	layout := "2006-01-02"
	t, err := time.Parse(layout, in.GetDate())
	if err != nil {
		log.Error(err)
	}

	lines := in.GetLines()
	for _, line := range lines {

		a := line.GetAccountname()
		acc, err := core.NewAccount(a, a)
		if err != nil {
			log.Error(err)
		}

		s, err := core.NewSplit(t, txn.Description, []*core.Account{acc}, aud, big.NewInt(line.GetAmount()))
		if err != nil {
			log.Error(err)
		}

		err = txn.AppendSplit(s)
		if err != nil {
			log.Error(err)
		}

	}

	s.ld.Insert(txn)

	return &pb.TransactionResponse{Message: "Accepted"}, nil
}

func (s *LedgerServer) NodeVersion(ctx context.Context, in *pb.VersionRequest) (*pb.VersionResponse, error) {
	log.Printf("Received Version Request: %s", in)
	return &pb.VersionResponse{Message: version.Version}, nil
}
