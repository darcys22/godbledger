package rpc

import (
	"context"
	"math/big"
	"strconv"
	"time"

	"github.com/darcys22/godbledger/godbledger/core"
	"github.com/darcys22/godbledger/godbledger/ledger"
	"github.com/darcys22/godbledger/godbledger/version"

	pb "github.com/darcys22/godbledger/proto"
)

type LedgerServer struct {
	pb.UnimplementedTransactorServer
	ld *ledger.Ledger
}

func (s *LedgerServer) AddTransaction(ctx context.Context, in *pb.TransactionRequest) (*pb.TransactionResponse, error) {
	log.WithField("Request", in).Info("Received New Add Transaction Request")

	usr, err := core.NewUser("MainUser")
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

		b := line.GetCurrency()
		curr, err := core.NewCurrency(b, 2)
		if err != nil {
			log.Error(err)
		}

		s, err := core.NewSplit(t, txn.Description, []*core.Account{acc}, curr, big.NewInt(line.GetAmount()))
		if err != nil {
			log.Error(err)
		}

		err = txn.AppendSplit(s)
		if err != nil {
			log.Error(err)
		}

	}

	response, err := s.ld.Insert(txn)
	if err != nil {
		log.Error(err)
	}

	return &pb.TransactionResponse{Message: response}, nil
}

func (s *LedgerServer) DeleteTransaction(ctx context.Context, in *pb.DeleteRequest) (*pb.TransactionResponse, error) {
	log.WithField("Request", in).Info("Received New Delete Request")
	s.ld.Delete(in.GetIdentifier())

	return &pb.TransactionResponse{Message: "Accepted"}, nil
}

func (s *LedgerServer) VoidTransaction(ctx context.Context, in *pb.DeleteRequest) (*pb.TransactionResponse, error) {
	log.WithField("Request", in).Info("Received New Void Request")

	usr, err := core.NewUser("MainUser")
	if err != nil {
		log.Error(err)
	}

	message := "Accepted"
	err = s.ld.Void(in.GetIdentifier(), usr)
	if err != nil {
		message = err.Error()
	}

	return &pb.TransactionResponse{Message: message}, nil
}

func (s *LedgerServer) NodeVersion(ctx context.Context, in *pb.VersionRequest) (*pb.VersionResponse, error) {
	log.WithField("Request", in).Info("Received New Version Request")
	return &pb.VersionResponse{Message: version.Version}, nil
}

func (s *LedgerServer) AddTag(ctx context.Context, in *pb.TagRequest) (*pb.TransactionResponse, error) {
	log.WithField("Request", in).Info("Received New Add Tag Request")

	s.ld.InsertTag(in.GetAccount(), in.GetTag())

	return &pb.TransactionResponse{Message: "Accepted"}, nil
}

func (s *LedgerServer) DeleteTag(ctx context.Context, in *pb.DeleteTagRequest) (*pb.TransactionResponse, error) {
	log.WithField("Request", in).Info("Received New Delete Tag Request")
	s.ld.DeleteTag(in.GetAccount(), in.GetTag())

	return &pb.TransactionResponse{Message: "Accepted"}, nil
}

func (s *LedgerServer) AddCurrency(ctx context.Context, in *pb.CurrencyRequest) (*pb.TransactionResponse, error) {
	log.WithField("Request", in).Info("Received New Add Currency Request")

	curr, err := core.NewCurrency(in.GetCurrency(), int(in.GetDecimals()))
	if err != nil {
		log.Error(err)
	}

	err = s.ld.InsertCurrency(curr)
	if err != nil {
		log.Error(err)
	}

	return &pb.TransactionResponse{Message: "Accepted"}, nil
}

func (s *LedgerServer) DeleteCurrency(ctx context.Context, in *pb.DeleteCurrencyRequest) (*pb.TransactionResponse, error) {
	log.WithField("Request", in).Info("Received New Delete Currency Request")
	s.ld.DeleteCurrency(in.GetCurrency())

	return &pb.TransactionResponse{Message: "Accepted"}, nil
}

func (s *LedgerServer) GetTB(ctx context.Context, in *pb.TBRequest) (*pb.TBResponse, error) {
	log.WithField("Request", in).Info("Received New Get Trial Balance Request")
	response := pb.TBResponse{}

	querydate, err := time.Parse("2006-01-02", in.Date)
	if err != nil {
		return &response, err
	}
	accounts, err := s.ld.GetTB(querydate)
	if err != nil {
		return &response, err
	}

	log.Debug("Building TB Response")
	for _, account := range *accounts {
		log.Debugf("Account: %s", account.Account)
		amt := strconv.Itoa(account.Amount)
		if len(amt) > account.Decimals {
			amt = amt[:len(amt)-account.Decimals] + "." + amt[len(amt)-account.Decimals:]
		} else {
			prefix := "0."
			for i := 1; i <= account.Decimals-len(amt); i++ {
				prefix = prefix + "0"
			}
			amt = prefix + amt
		}
		response.Lines = append(response.Lines,
			&pb.TBLine{
				Accountname: account.Account,
				Amount:      int64(account.Amount),
				Tags:        account.Tags,
				Currency:    account.Currency,
				Decimals:    int64(account.Decimals),
				AmountStr:   amt,
			})
	}

	return &response, nil
}

func (s *LedgerServer) GetListing(ctx context.Context, in *pb.ReportRequest) (*pb.ListingResponse, error) {
	log.WithField("Request", in).Info("Received New Get Listing Request")
	response := pb.ListingResponse{}

	startdate, err := time.Parse("2006-01-02", in.Startdate)
	if err != nil {
		return &response, err
	}
	enddate, err := time.Parse("2006-01-02", in.Date)
	if err != nil {
		return &response, err
	}
	txns, err := s.ld.GetListing(startdate, enddate)
	if err != nil {
		return &response, err
	}

	log.Debug("Building Listing Response")

	for _, txn := range *txns {
		splits := []*pb.LineItem{}
		date := ""

		if len(txn.Splits) > 0 {
			date = txn.Splits[0].Date.Format("2006-01-02 15:04:05")
			for _, split := range txn.Splits {
				splits = append(splits,
					&pb.LineItem{
						Accountname: split.Accounts[0].Name,
						Description: string(split.Description),
						Currency:    split.Currency.Name,
						Amount:      split.Amount.Int64(),
					})
			}
		} else {
			date = txn.Postdate.Format("2006-01-02 15:04:05")
		}
		response.Transactions = append(response.Transactions,
			&pb.Transaction{
				Date:        date,
				Description: string(txn.Description),
				Lines:       splits,
			})
	}

	return &response, nil
}
