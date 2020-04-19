package rpc

import (
	"context"
	"math/big"
	"time"

	"github.com/darcys22/godbledger/godbledger/core"
	"github.com/darcys22/godbledger/godbledger/ledger"
	"github.com/darcys22/godbledger/godbledger/version"

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
	log.Info("Received Version Request: %s", in)
	return &pb.VersionResponse{Message: version.Version}, nil
}

func (s *LedgerServer) DeleteTransaction(ctx context.Context, in *pb.DeleteRequest) (*pb.TransactionResponse, error) {
	log.Info("Received New Delete Request")
	s.ld.Delete(in.GetIdentifier())

	return &pb.TransactionResponse{Message: "Accepted"}, nil
}

func (s *LedgerServer) AddTag(ctx context.Context, in *pb.TagRequest) (*pb.TransactionResponse, error) {
	log.Info("Received New Add Tag Request")

	s.ld.InsertTag(in.GetAccount(), in.GetTag())

	return &pb.TransactionResponse{Message: "Accepted"}, nil
}

func (s *LedgerServer) DeleteTag(ctx context.Context, in *pb.DeleteTagRequest) (*pb.TransactionResponse, error) {
	log.Info("Received New Delete Request")
	s.ld.DeleteTag(in.GetAccount(), in.GetTag())

	return &pb.TransactionResponse{Message: "Accepted"}, nil
}

func (s *LedgerServer) GetTB(ctx context.Context, in *pb.TBRequest) (*pb.TBResponse, error) {
	log.Info("Received New TB Request")
	accounts, err := s.ld.GetTB(time.Now())

	//log.Debug(accounts)

	response := pb.TBResponse{}

	for _, account := range *accounts {
		response.Lines = append(response.Lines,
			&pb.TBLine{
				Accountname: account.Account,
				Amount:      int64(account.Amount),
				Tags:        account.Tags,
			})
	}

	return &response, err
}
