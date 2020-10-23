// +build none

/*
The test command is called to input bulk data

Usage: go run utils/ci.go <command> <command flags/arguments>

Available commands are:

   income                              -- submits bulk income transactions
   expense                             -- submits bulk expense transactions

*/
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"time"

	pb "github.com/darcys22/godbledger/proto"
	"google.golang.org/grpc"
)

const (
	address = "localhost:50051"
)

func main() {
	log.SetFlags(log.Lshortfile)

	if _, err := os.Stat(filepath.Join("utils", "testdata.go")); os.IsNotExist(err) {
		log.Fatal("this script must be run from the root of the repository")
	}
	if len(os.Args) < 2 {
		log.Fatal("need subcommand as first argument")
	}
	switch os.Args[1] {
	case "income":
		doInstall("utils/generatedIncome.json")
	case "expense":
		doInstall("utils/generatedExpenses.json")
	default:
		log.Fatal("unknown command ", os.Args[1])
	}
}

func doInstall(jsonfile string) {

	// read file
	data, err := ioutil.ReadFile(jsonfile)
	if err != nil {
		fmt.Print(err)
	}

	// define data structure
	type Transactions struct {
		Description string
		Date        string
		Account     string
		Balance     int64
	}

	// json data
	var obj []Transactions

	// unmarshall it
	err = json.Unmarshal(data, &obj)
	if err != nil {
		fmt.Println("error unmarshaling:", err)
	}

	for i := 0; i < len(obj); i++ {
		date, err := time.Parse("2006-01-02T15:04:05 Z07:00", obj[i].Date)
		if err != nil {
			fmt.Println("error parsing date:", err)
		}
		fmt.Println(date)

		transactionLines := make([]Account, 2)

		line1Account := obj[i].Account
		line1Desc := obj[i].Description[:20]
		line1Amount := big.NewRat(obj[i].Balance, 1)

		transactionLines[0] = Account{
			Name:        line1Account,
			Description: line1Desc,
			Balance:     line1Amount,
			Currency:    "USD",
		}

		line2Account := "Assets:Checking"
		line2Desc := obj[i].Description[:20]
		line2Amount := big.NewRat(obj[i].Balance*-1, 1)

		transactionLines[1] = Account{
			Name:        line2Account,
			Description: line2Desc,
			Balance:     line2Amount,
			Currency:    "USD",
		}

		req := &Transaction{
			Date:           date,
			Payee:          obj[i].Description[:20],
			AccountChanges: transactionLines,
		}

		err = Send(req)
		if err != nil {
			log.Fatalf("could not send: %v", err)
		}
	}

}

// Account holds the name and balance
type Account struct {
	Name        string
	Description string
	Currency    string
	Balance     *big.Rat
}

type sortAccounts []*Account

func (s sortAccounts) Len() int      { return len(s) }
func (s sortAccounts) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type sortAccountsByName struct{ sortAccounts }

func (s sortAccountsByName) Less(i, j int) bool {
	return s.sortAccounts[i].Name < s.sortAccounts[j].Name
}

// Transaction is the basis of a ledger. The ledger holds a list of transactions.
// A Transaction has a Payee, Date (with no time, or to put another way, with
// hours,minutes,seconds values that probably doesn't make sense), and a list of
// Account values that hold the value of the transaction for each account.
type Transaction struct {
	Payee          string
	Date           time.Time
	AccountChanges []Account
}

type sortTransactions []*Transaction

func (s sortTransactions) Len() int      { return len(s) }
func (s sortTransactions) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type sortTransactionsByDate struct{ sortTransactions }

func (s sortTransactionsByDate) Less(i, j int) bool {
	return s.sortTransactions[i].Date.Before(s.sortTransactions[j].Date)
}

func Send(t *Transaction) error {

	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewTransactorClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	transactionLines := make([]*pb.LineItem, 2)

	for i, accChange := range t.AccountChanges {
		transactionLines[i] = &pb.LineItem{
			Accountname: accChange.Name,
			Description: accChange.Description,
			Amount:      accChange.Balance.Num().Int64(),
			Currency:    accChange.Currency,
		}
	}

	req := &pb.TransactionRequest{
		Date:        t.Date.Format("2006-01-02"),
		Description: t.Payee,
		Lines:       transactionLines,
	}
	r, err := client.AddTransaction(ctx, req)
	if err != nil {
		log.Fatalf("Could not send transaction: %v", err)
	}
	log.Printf("Response: %s", r.GetMessage())
	return nil
}
