// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// Package tests implements execution of Ethereum JSON tests.
package tests

import (
	"encoding/json"
)

type BasicTest struct {
	json btJSON
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (t *BasicTest) UnmarshalJSON(in []byte) error {
	return json.Unmarshal(in, &t.json)
}

type btJSON struct {
	Network string `json:"network"`
}

func (t *BasicTest) Run(snapshotter bool) error {
	// Run Test here
	//if err = t.validatePostState(newDB); err != nil {
	//return fmt.Errorf("post state validation failed: %v", err)
	//}
	return nil
}

//func (t *BlockTest) validatePostState(statedb *state.StateDB) error {
//// validate post state accounts in test file against what we have in state db
//for addr, acct := range t.json.Post {
//// address is indirectly verified by the other fields, as it's the db key
//code2 := statedb.GetCode(addr)
//balance2 := statedb.GetBalance(addr)
//nonce2 := statedb.GetNonce(addr)
//if !bytes.Equal(code2, acct.Code) {
//return fmt.Errorf("account code mismatch for addr: %s want: %v have: %s", addr, acct.Code, hex.EncodeToString(code2))
//}
//if balance2.Cmp(acct.Balance) != 0 {
//return fmt.Errorf("account balance mismatch for addr: %s, want: %d, have: %d", addr, acct.Balance, balance2)
//}
//if nonce2 != acct.Nonce {
//return fmt.Errorf("account nonce mismatch for addr: %s want: %d have: %d", addr, acct.Nonce, nonce2)
//}
//}
//return nil
//}
