package core

type TBAccount struct {
	Account  string   `json:"Account"`
	Amount   int      `json:"Amount"`
	Tags     []string `json:"Tags"`
	Currency string   `json:"Currency"`
	Decimals int      `json:"Decimals"`
}
