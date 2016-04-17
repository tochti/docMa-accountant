package accountant

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tochti/docMa-handler/accountingData"
)

func Test_FindAccountingTxsWithoutVouchers(t *testing.T) {
	db := initGorpConn(t)
	createVouchersInDB(t, db)

	// remove db entry

	a, err := FindAccountingTxsWithoutVouchers(AccountingTxsFile)
	if err != nil {
		t.Fatal(err)
	}

	eAccTx := accountingData.AccountingData{
		DocDate:          time.Date(2013, 8, 29, 0, 0, 0, 0, time.UTC),
		DateOfEntry:      time.Date(2013, 9, 1, 0, 0, 0, 0, time.UTC),
		DocNumberRange:   "B",
		DocNumber:        "6",
		PostingText:      "Lastschrift Strato",
		AmountPosted:     7.99,
		DebitAccount:     71003,
		CreditAccount:    1210,
		TaxCode:          0,
		CostUnit1:        "",
		CostUnit2:        "",
		AmountPostedEuro: 7.99,
		Currency:         "EUR",
	}

	accTx := a[0]
	assert.Equal(t, eAccTx, accTx)

}
