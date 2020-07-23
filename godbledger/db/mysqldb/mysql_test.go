//
//  go-unit-test-sql
//
//  Copyright © 2020. All rights reserved.
//  https://medium.com/@bismobaruno/unit-test-sql-in-golang-5af19075e68e
//

package mysqldb

import (
	//"database/sql"
	//"log"
	"testing"

	//"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestValidateConnectionString(t *testing.T) {

	//Test Regular string with no params
	validatedString, err := ValidateConnectionString("godbledger:password@tcp(192.168.1.98:3306)/ledger")
	assert.Nil(t, err)
	if assert.NotNil(t, validatedString, "Connection String with no params") {
		assert.Equal(t, "godbledger:password@tcp(192.168.1.98:3306)/ledger?parseTime=true&charset=utf8", validatedString)
	}

	//Test Same String with a param to ensure no duplication
	validatedString, err = ValidateConnectionString("godbledger:password@tcp(192.168.1.98:3306)/ledger?parseTime=true")
	assert.Nil(t, err)
	if assert.NotNil(t, validatedString, "Connection String with param") {
		assert.Equal(t, "godbledger:password@tcp(192.168.1.98:3306)/ledger?parseTime=true&charset=utf8", validatedString)
	}

	//Test empty string to ensure error
	nilString, err := ValidateConnectionString("")
	if assert.NotNil(t, err, "Nil connection string") {
		assert.Equal(t, "Connection string not provided", err.Error())
	}
	assert.Equal(t, "", nilString)
}
