package rpc

import (
	"context"
	"math/big"
	"strconv"
	"time"

	"github.com/darcys22/godbledger/proto/transaction"

	"github.com/darcys22/godbledger/godbledger/core"
	"github.com/darcys22/godbledger/godbledger/ledger"
	"github.com/darcys22/godbledger/godbledger/version"
)

type LedgerServer struct {
	transaction.UnimplementedTransactorServer
	ld *ledger.Ledger
}

func (s *LedgerServer) AddTransaction(ctx context.Context, in *transaction.TransactionRequest) (*transaction.TransactionResponse, error) {
	log.WithField("Request", in).Info("Received New Add Transaction Request")

	usr, err := core.NewUser("MainUser")
	if err != nil {
		log.Infof("Add Transaction error: %s", err.Error())
		return &transaction.TransactionResponse{}, err
	}

	txn, err := core.NewTransaction(usr)
	if err != nil {
		log.Infof("Add Transaction error: %s", err.Error())
		return &transaction.TransactionResponse{}, err
	}
	txn.Description = []byte(in.GetDescription())

	layout := "2006-01-02"
	t, err := time.Parse(layout, in.GetDate())
	if err != nil {
		log.Infof("Add Transaction error: %s", err.Error())
		return &transaction.TransactionResponse{}, err
	}

	lines := in.GetLines()
	for _, line := range lines {
		a := line.GetAccountname()
		acc, err := core.NewAccount(a, a)
		if err != nil {
			log.Infof("Add Transaction error: %s", err.Error())
			return &transaction.TransactionResponse{}, err
		}

		b := line.GetCurrency()
		curr, err := s.ld.GetCurrency(b)
		if err != nil {
			log.Infof("Add Transaction error: %s", err.Error())
			return &transaction.TransactionResponse{}, err
		}

		s, err := core.NewSplit(t, txn.Description, []*core.Account{acc}, curr, big.NewInt(line.GetAmount()))
		if err != nil {
			log.Infof("Add Transaction error: %s", err.Error())
			return &transaction.TransactionResponse{}, err
		}

		err = txn.AppendSplit(s)
		if err != nil {
			log.Infof("Add Transaction error: %s", err.Error())
			return &transaction.TransactionResponse{}, err
		}
	}

	response, err := s.ld.Insert(txn)
	if err != nil {
		log.Infof("Add Transaction error: %s", err.Error())
		return &transaction.TransactionResponse{}, err
	}

	return &transaction.TransactionResponse{Message: response}, nil
}

func (s *LedgerServer) DeleteTransaction(ctx context.Context, in *transaction.DeleteRequest) (*transaction.TransactionResponse, error) {
	log.WithField("Request", in).Info("Received New Delete Request")
	s.ld.Delete(in.GetIdentifier())

	return &transaction.TransactionResponse{Message: "Accepted"}, nil
}

func (s *LedgerServer) VoidTransaction(ctx context.Context, in *transaction.DeleteRequest) (*transaction.TransactionResponse, error) {
	log.WithField("Request", in).Info("Received New Void Request")

	usr, err := core.NewUser("MainUser")
	if err != nil {
		log.Infof("Void Transaction error: %s", err.Error())
		return &transaction.TransactionResponse{}, err
	}

	message := "Accepted"
	err = s.ld.Void(in.GetIdentifier(), usr)
	if err != nil {
		log.Infof("Void Transaction error: %s", err.Error())
		return &transaction.TransactionResponse{}, err
	}

	return &transaction.TransactionResponse{Message: message}, nil
}

func (s *LedgerServer) NodeVersion(ctx context.Context, in *transaction.VersionRequest) (*transaction.VersionResponse, error) {
	log.WithField("Request", in).Info("Received New Version Request")
	return &transaction.VersionResponse{Message: version.Version}, nil
}

func (s *LedgerServer) AddTag(ctx context.Context, in *transaction.AccountTagRequest) (*transaction.TransactionResponse, error) {
	log.WithField("Request", in).Info("Received New Add Tag Request")

	tags := in.GetTag()
	for i := 0; i < len(tags); i++ {
		s.ld.InsertTag(in.GetAccount(), tags[i])
	}

	return &transaction.TransactionResponse{Message: "Accepted"}, nil
}

func (s *LedgerServer) DeleteTag(ctx context.Context, in *transaction.DeleteAccountTagRequest) (*transaction.TransactionResponse, error) {
	log.WithField("Request", in).Info("Received New Delete Tag Request")

	tags := in.GetTag()
	for i := 0; i < len(tags); i++ {
		s.ld.DeleteTag(in.GetAccount(), tags[i])
	}

	return &transaction.TransactionResponse{Message: "Accepted"}, nil
}

func (s *LedgerServer) AddAccount(ctx context.Context, in *transaction.AccountTagRequest) (*transaction.TransactionResponse, error) {
	log.WithField("Request", in).Info("Received New Add Account Request")

	accountRequested := in.GetAccount()
	err := s.ld.InsertAccount(accountRequested)
	if err != nil {
		log.Infof("Add Account error: %s", err.Error())
		return &transaction.TransactionResponse{}, err
	}

	tags := in.GetTag()
	for i := 0; i < len(tags); i++ {
		s.ld.InsertTag(accountRequested, tags[i])
	}

	return &transaction.TransactionResponse{Message: "Accepted"}, nil
}

func (s *LedgerServer) DeleteAccount(ctx context.Context, in *transaction.DeleteAccountTagRequest) (*transaction.TransactionResponse, error) {
	log.WithField("Request", in).Info("Received New Delete Account Request")

	accountRequested := in.GetAccount()

	tags := in.GetTag()
	for i := 0; i < len(tags); i++ {
		s.ld.DeleteTag(accountRequested, tags[i])
	}

	err := s.ld.DeleteAccount(accountRequested)
	if err != nil {
		log.Infof("Delete Account error: %s", err.Error())
		return &transaction.TransactionResponse{}, err
	}

	return &transaction.TransactionResponse{Message: "Accepted"}, nil
}

func (s *LedgerServer) AddCurrency(ctx context.Context, in *transaction.CurrencyRequest) (*transaction.TransactionResponse, error) {
	log.WithField("Request", in).Info("Received New Add Currency Request")

	curr, err := core.NewCurrency(in.GetCurrency(), int(in.GetDecimals()))
	if err != nil {
		log.Infof("Add Currency error: %s", err.Error())
		return &transaction.TransactionResponse{}, err
	}

	err = s.ld.InsertCurrency(curr)
	if err != nil {
		log.Infof("Add Currency error: %s", err.Error())
		return &transaction.TransactionResponse{}, err
	}
	return &transaction.TransactionResponse{Message: "Accepted"}, nil
}

func (s *LedgerServer) DeleteCurrency(ctx context.Context, in *transaction.DeleteCurrencyRequest) (*transaction.TransactionResponse, error) {
	log.WithField("Request", in).Info("Received New Delete Currency Request")
	s.ld.DeleteCurrency(in.GetCurrency())

	return &transaction.TransactionResponse{Message: "Accepted"}, nil
}

func (s *LedgerServer) ReconcileTransactions(ctx context.Context, in *transaction.ReconciliationRequest) (*transaction.TransactionResponse, error) {
	log.WithField("Request", in).Info("Received New Reconciliation Request")
	response := transaction.TransactionResponse{}
	reconciliationID, err := s.ld.ReconcileTransactions(in.GetSplitID())

	if err != nil {
		log.Infof("Reconcile Transactions error: %s", err.Error())
		return &transaction.TransactionResponse{}, err
	}

	log.WithField("reconciliation ID", reconciliationID).Debug("Created Reconciliation")

	response.Message = reconciliationID
	return &response, nil
}

func (s *LedgerServer) GetTB(ctx context.Context, in *transaction.TBRequest) (*transaction.TBResponse, error) {
	log.WithField("Request", in).Info("Received New Get Trial Balance Request")
	response := transaction.TBResponse{}

	querydate, err := time.Parse("2006-01-02", in.Date)
	if err != nil {
		log.Infof("Get Trial Balance error: %s", err.Error())
		return &transaction.TBResponse{}, err
	}
	accounts, err := s.ld.GetTB(querydate)
	if err != nil {
		log.Infof("Get Trial Balance error: %s", err.Error())
		return &transaction.TBResponse{}, err
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
			&transaction.TBLine{
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

func (s *LedgerServer) GetListing(ctx context.Context, in *transaction.ReportRequest) (*transaction.ListingResponse, error) {
	log.WithField("Request", in).Info("Received New Get Listing Request")
	response := transaction.ListingResponse{}

	startdate, err := time.Parse("2006-01-02", in.Startdate)
	if err != nil {
		log.Infof("Get Listing error: %s", err.Error())
		return &transaction.ListingResponse{}, err
	}
	enddate, err := time.Parse("2006-01-02", in.Date)
	if err != nil {
		log.Infof("Get Listing error: %s", err.Error())
		return &transaction.ListingResponse{}, err
	}
	txns, err := s.ld.GetListing(startdate, enddate)
	if err != nil {
		log.Infof("Get Listing error: %s", err.Error())
		return &transaction.ListingResponse{}, err
	}

	log.Debug("Building Listing Response")

	for _, txn := range *txns {
		splits := []*transaction.LineItem{}
		date := ""

		if len(txn.Splits) > 0 {
			date = txn.Splits[0].Date.Format("2006-01-02 15:04:05")
			for _, split := range txn.Splits {
				splits = append(splits,
					&transaction.LineItem{
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
			&transaction.Transaction{
				Date:        date,
				Description: string(txn.Description),
				Lines:       splits,
			})
	}

	return &response, nil
}

func (s *LedgerServer) AddTransactionFeedAccount(ctx context.Context, in *transaction.TransactionFeedAccountRequest) (*transaction.TransactionResponse, error) {
	log.WithField("Request", in).Info("Received New Add Transaction Feed Account Request")

	err := s.ld.InsertFeedAccount(in.GetName(), in.GetCurrency())
	if err != nil {
		log.Infof("Add Account error: %s", err.Error())
		return &transaction.TransactionResponse{}, err
	}

  s.ld.InsertTag(in.GetName(), "feed")

	return &transaction.TransactionResponse{Message: "Accepted"}, nil
}

func (s *LedgerServer) AddTransactionFeed(ctx context.Context, in *transaction.TransactionFeedRequest) (*transaction.TransactionResponse, error) {
	log.WithField("Request", in).Info("Received New Add Feed Request")

	usr, err := core.NewUser("MainUser")
	if err != nil {
		log.Infof("Add Transaction Feed error 1: %s", err.Error())
		return &transaction.TransactionResponse{}, err
	}

	account, currency, err := s.ld.GetAccount(in.GetAccount())
  if err != nil {
    log.Infof("Add Transaction Feed error 2: %s", err.Error())
    return &transaction.TransactionResponse{}, err
  }

	lines := in.GetLines()
	for _, line := range lines {
    txn, err := core.NewTransaction(usr)
    if err != nil {
      log.Infof("Add Transaction Feed error 3: %s", err.Error())
      return &transaction.TransactionResponse{}, err
    }
    txn.Description = []byte(line.GetHash())

    layout := "2006-01-02"
    t, err := time.Parse(layout, line.GetDate())
    if err != nil {
      log.Infof("Add Transaction Feed error 4: %s", err.Error())
      return &transaction.TransactionResponse{}, err
    }

		split, err := core.NewSplit(t, []byte(line.GetDescription()), []*core.Account{&account}, &currency, big.NewInt(line.GetAmount()))
		if err != nil {
			log.Infof("Add Transaction Feed error 5: %s", err.Error())
			return &transaction.TransactionResponse{}, err
		}

		err = txn.AppendSplit(split)
		if err != nil {
			log.Infof("Add Transaction Feed error 6: %s", err.Error())
			return &transaction.TransactionResponse{}, err
		}
    _ , err = s.ld.Insert(txn)
    if err != nil {
      log.Infof("Add Transaction Feed error 7: %s", err.Error())
      return &transaction.TransactionResponse{}, err
    }
	}


	return &transaction.TransactionResponse{Message: "Accepted"}, nil
}
