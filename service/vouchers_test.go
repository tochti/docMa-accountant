package accountant

import (
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tochti/docMa-handler/accountingData"
	"github.com/tochti/docMa-handler/docs"
	"github.com/tochti/gin-gum/gumspecs"
	"gopkg.in/gorp.v1"
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

func createVouchersInDB(t *testing.T, db *gorp.DbMap) {
	docs.AddTables(db)

	err := db.DropTablesIfExists()
	if err != nil {
		t.Fatal(err)
	}
	err = db.CreateTablesIfNotExists()
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.Open(AccountingTxsFile)
	if err != nil {
		t.Fatal(err)
	}
	r := NewReader(f)

	for n := 0; ; n++ {
		a, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}

			t.Fatal(err)
		}

		d := docs.DocAccountData{
			PeriodFrom:    a.DocDate.Add(-24 * time.Hour),
			PeriodTo:      a.DocDate.Add(24 * time.Hour),
			AccountNumber: a.CreditAccount,
		}
		err = db.Insert(&d)
		if err != nil {
			t.Fatal(err)
		}

		docNumber := ""
		if a.DocNumber != "" {
			docNumber = a.DocNumber
		}
		dn := docs.DocNumber{
			DocID:  d.DocID,
			Number: docNumber,
		}
		err = db.Insert(&dn)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func initGorpConn(t *testing.T) *gorp.DbMap {
	mysql := gumspecs.ReadMySQL()

	c, err := mysql.DB()
	if err != nil {
		t.Fatal(err)
	}

	return &gorp.DbMap{
		Db: c,
		Dialect: gorp.MySQLDialect{
			"InnonDB",
			"UTF8",
		},
	}
}