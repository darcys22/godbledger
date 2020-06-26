package core

import (
	"time"

	"github.com/rs/xid"
	"math/big"
)

type User struct {
	Id   string
	Name string
}

func NewUser(name string) (*User, error) {
	guid := xid.New()

	usr := &User{guid.String(), name}
	return usr, nil

}

type Currency struct {
	Name     string
	Decimals int
}

func NewCurrency(name string, decimals int) (*Currency, error) {

	cur := &Currency{name, decimals}
	return cur, nil

}

type Account struct {
	Code string
	Name string
}

func NewAccount(code, name string) (*Account, error) {
	acc := &Account{code, name}
	return acc, nil
}

type Transaction struct {
	Id          string
	Postdate    time.Time
	Poster      *User
	Description []byte
	Splits      []*Split
}

func NewTransaction(usr *User) (*Transaction, error) {
	guid := xid.New()
	txn := &Transaction{guid.String(), time.Now(), usr, []byte{}, []*Split{}}
	return txn, nil
}

func ReverseTransaction(originalTxn *Transaction, usr *User) (*Transaction, error) {
	guid := xid.New()
	txn := &Transaction{guid.String(), time.Now(), usr, []byte{}, []*Split{}}

	for _, split := range originalTxn.Splits {
		newSplt, err := NewSplit(split.Date, split.Description, split.Accounts, split.Currency, big.NewInt(0).Mul(big.NewInt(-1), split.Amount))
		if err != nil {
			return nil, err
		}
		txn.AppendSplit(newSplt)
	}
	return txn, nil
}

func (txn *Transaction) AppendSplit(spl *Split) error {
	txn.Splits = append(txn.Splits, spl)
	return nil
}

func (txn *Transaction) Balance() (*big.Int, bool) {
	valid := true
	if len(txn.Splits) < 1 {
		valid = false
	}
	total := big.NewInt(0)
	for _, elem := range txn.Splits {
		total.Add(total, elem.Amount)
	}

	if total.Cmp(big.NewInt(0)) != 0 {
		valid = false
	}
	return total, valid
}

type Split struct {
	Id          string
	Date        time.Time
	Description []byte
	Accounts    []*Account
	Currency    *Currency
	Amount      *big.Int
}

func NewSplit(date time.Time, desc []byte, accs []*Account, cur *Currency, amt *big.Int) (*Split, error) {
	guid := xid.New()
	spl := &Split{guid.String(), date, desc, accs, cur, amt}
	return spl, nil
}
