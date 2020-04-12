package core

import ()

type Tag struct {
	Name     string       `json:"Name"`
	Total    int          `json:"Total"`
	Accounts []PDFAccount `json:"Accounts"`
}

type PDFAccount struct {
	Account string `json:"Account"`
	Amount  int    `json:"Amount"`
}

var Reporteroutput struct {
	Data      []Tag `json:"Tags"`
	Profit    int   `json:"Profit"`
	NetAssets int   `json:"NetAssets"`
}
