package evaluators

import (
	"context"
	"errors"
	"fmt"
	"time"

	pb "github.com/darcys22/godbledger/proto"

	"github.com/darcys22/godbledger/tests/types"

	"google.golang.org/grpc"
)

// SingleTransaction submits a single transaction to the server and expects no errors as a response
var SingleTransaction = types.Evaluator{
	Name:       "single_transaction",
	Evaluation: singleTransaction,
}

func singleTransaction(conns ...*grpc.ClientConn) error {
	client := pb.NewTransactorClient(conns[0])

	date, _ := time.Parse("2006-01-02", "2011-03-15")
	desc := "Whole Food Market"

	transactionLines := make([]*pb.LineItem, 2)

	transactionLines[0] = &pb.LineItem{
		Accountname: "Expenses:Groceries",
		Description: "Groceries",
		Amount:      7500,
		Currency:    "USD",
	}

	transactionLines[1] = &pb.LineItem{
		Accountname: "Assets:Checking",
		Description: "Groceries",
		Amount:      -7500,
		Currency:    "USD",
	}

	req := &pb.TransactionRequest{
		Date:        date.Format("2006-01-02"),
		Description: desc,
		Lines:       transactionLines,
		Signature:   "test",
	}
	_, err := client.AddTransaction(context.Background(), req)
	if err != nil {
		return err
	}

	res, err := client.GetTB(context.Background(), &pb.TBRequest{Date: time.Now().Format("2006-01-02")})
	if err != nil {
		return err
	}

	// Initialise a variable to check that the trial balance balances
	balance := int64(0)
	// Check to ensure the Trial Balance Matches.
	for i := 0; i < len(res.Lines); i++ {
		balance += res.Lines[i].Amount
		switch res.Lines[i].Accountname {
		case "Assets:Checking":
			if res.Lines[i].Amount != int64(-7500) {
				return errors.New("Trial Balance Checking Account Incorrect")
			}
		case "Expenses:Groceries":
			if res.Lines[i].Amount != int64(7500) {
				return errors.New("Trial Balance Groceries Account Incorrect")
			}
		default:
			return errors.New(fmt.Sprintf("Unknown Account %s", res.Lines[i].Accountname))
		}
	}

	return nil
}
