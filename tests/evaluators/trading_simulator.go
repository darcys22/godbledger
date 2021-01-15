package evaluators

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/darcys22/godbledger/proto/transaction"
	"github.com/darcys22/godbledger/tests/types"

	"google.golang.org/grpc"
)

const (
	address       = "localhost:50051"
	iterationDays = 365
	sdBPS         = 500
	decimalsBPS   = 10000
	tradProb      = 10
	startPrice    = 100.00
)

// Account holds the name and balance
type Trade struct {
	amount int
	price  float64
}

// TradingSimulator submits a random transactions to the server
var TradingSimulator = types.Evaluator{
	Name:       "Trading Simulator",
	Evaluation: tradingSimulator,
}

func tradingSimulator(conns ...*grpc.ClientConn) error {

	client := transaction.NewTransactorClient(conns[0])

	ctx := context.Background()

	//rand.Seed(42)
	rand.Seed(time.Now().UTC().UnixNano())

	positions := []Trade{}
	price := startPrice

	for day := 1; day <= iterationDays; day++ {
		price = price * (1 + math.Round(rand.NormFloat64()*sdBPS)/decimalsBPS)
		//fmt.Printf("Price: %.2f \n", price)
		if rand.Intn(10) == 1 {
			if rand.Intn(2) == 1 {
				amount := rand.Intn(100)
				positions = append(positions, Trade{amount, price})
				fmt.Printf("Buy: %d \n", amount)
				fmt.Printf("Unit Price: %.2f \n", price)
				fmt.Printf("Total Paid: %.2f \n\n", price*float64(amount))

				//Create the purchase transaction to send to the server
				date := "2011-03-15"
				desc := "Buy Purchase on dd mmm yyyy\n\n"
				desc += fmt.Sprintf("Buy: %d \n", amount)
				desc += fmt.Sprintf("Unit Price: %.2f \n", price)
				desc += fmt.Sprintf("Total Paid: %.2f \n\n", price*float64(amount))

				transactionLines := make([]*transaction.LineItem, 2)

				line1Account := "Assets:Cash"
				line1Desc := "Buy Purchase on dd mmm yyyy\n\n"
				line1Amount := int64(math.Round(price*100)) * int64(amount) * -1

				transactionLines[0] = &transaction.LineItem{
					Accountname: line1Account,
					Description: line1Desc,
					Amount:      line1Amount,
				}

				line2Account := "Assets:Crypto"
				line2Desc := "Buy Purchase on dd mmm yyyy\n\n"
				line2Amount := int64(math.Round(price*100)) * int64(amount)

				transactionLines[1] = &transaction.LineItem{
					Accountname: line2Account,
					Description: line2Desc,
					Amount:      line2Amount,
				}

				req := &transaction.TransactionRequest{
					Date:        date,
					Description: desc,
					Lines:       transactionLines,
				}
				r, err := client.AddTransaction(ctx, req)
				if err != nil {
					return fmt.Errorf("could not greet: %v", err)
				}
				log.Printf("Version: %s", r.GetMessage())
			} else {
				amount := rand.Intn(100)
				fmt.Printf("Sell %d \n", amount)
				pricepaid := 0.0
				purchased := 0
				for purchased < amount {
					if len(positions) > 0 {
						n := positions[0]
						positions = positions[1:]
						if n.amount >= (amount - purchased) {
							purchased += (amount - purchased)
							pricepaid += float64(amount) * n.price
							positions = append([]Trade{Trade{n.amount - amount, n.price}}, positions...)
						} else {
							purchased += n.amount
							pricepaid += float64(n.amount) * n.price
						}
					} else {
						amount = purchased
					}
				}
				fmt.Printf("Sell %d \n", amount)
				fmt.Printf("Unit Price: %.2f \n", price)
				fmt.Printf("Proceeds: %.2f \n", price*float64(amount))
				fmt.Printf("Sell %d \n", purchased)
				fmt.Printf("Profit: %.2f \n\n", float64(amount)*price-pricepaid)

				//Create the sale transaction to send to the server
				date := "2011-03-15"
				desc := "Sell order on dd mmm yyyy\n\n"
				desc += fmt.Sprintf("Sell %d \n", amount)
				desc += fmt.Sprintf("Unit Price: %.2f \n", price)
				desc += fmt.Sprintf("Proceeds: %.2f \n", price*float64(amount))
				desc += fmt.Sprintf("Sell %d \n", purchased)
				desc += fmt.Sprintf("Profit: %.2f \n\n", float64(amount)*price-pricepaid)

				transactionLines := make([]*transaction.LineItem, 3)

				line1Account := "Assets:Cash"
				line1Desc := "Sell order on dd mmm yyyy\n\n"
				line1Amount := int64(math.Round(price*100)) * int64(amount)

				transactionLines[0] = &transaction.LineItem{
					Accountname: line1Account,
					Description: line1Desc,
					Amount:      line1Amount,
				}

				line2Account := "Assets:Crypto"
				line2Desc := "Sell order on dd mmm yyyy\n\n"
				line2Amount := int64(math.Round(pricepaid * 100))

				transactionLines[1] = &transaction.LineItem{
					Accountname: line2Account,
					Description: line2Desc,
					Amount:      line2Amount,
				}

				line3Account := "Revenue:Trading"
				line3Desc := "Sell order on dd mmm yyyy\n\n"
				line3Amount := int64(math.Round(price*100))*int64(amount) - int64(math.Round(pricepaid*100))

				transactionLines[2] = &transaction.LineItem{
					Accountname: line3Account,
					Description: line3Desc,
					Amount:      line3Amount,
				}

				req := &transaction.TransactionRequest{
					Date:        date,
					Description: desc,
					Lines:       transactionLines,
				}
				r, err := client.AddTransaction(ctx, req)
				if err != nil {
					return fmt.Errorf("could not greet: %v", err)
				}
				log.Printf("Version: %s", r.GetMessage())
			}
		}
	}

	return nil
}
