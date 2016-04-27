package accountantService

import (
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tochti/docMa-accountant/accountantService/accountingTxsFileReader"
	"github.com/tochti/docMa-handler/accountingData"
	"github.com/tochti/docMa-handler/docs"
	"github.com/tochti/gin-gum/gumspecs"
	"gopkg.in/gorp.v1"
)

var (
	TestData   = "./test_data"
	TestDBName = "testing"
)

func Test_FindVouchersByID(t *testing.T) {

	cases := map[string]struct {
		Number   string
		Expected []docs.Doc
	}{
		"default": {
			Number: "B7",
			Expected: []docs.Doc{
				docs.Doc{
					ID:   1,
					Name: "lumpi",
				},
			},
		},
	}

	db := setupTestDB(t)
	srv := Service{
		DB: db,
	}
	for k, tc := range cases {

		for i, expect := range tc.Expected {
			resetTestTables(t, db)
			err := db.Insert(&tc.Expected[i], &docs.DocNumber{tc.Expected[i].ID, tc.Number})
			if err != nil {
				t.Fatal(err)
			}

			dl, err := srv.FindVouchers(tc.Number, -1, zeroDate())
			if err != nil {
				fmt.Errorf("%v: %v", k, err)
			}

			if expect.ID != dl[i].ID ||
				expect.Name != dl[i].Name {
				fmt.Errorf("%v: Expect %v was %v", k, expect, dl[i])
			}
		}
	}
}

func Test_FindVouchersByAccountNumber(t *testing.T) {
	cases := map[string]struct {
		AccountNumber int
		Date          time.Time
		Expected      []docs.Doc
	}{
		"default": {
			AccountNumber: 1400,
			Date:          time.Now(),
			Expected: []docs.Doc{
				docs.Doc{
					ID:   1,
					Name: "lumpi",
				},
			},
		},
	}

	db := setupTestDB(t)
	srv := Service{
		DB: db,
	}
	for k, tc := range cases {

		for i, expect := range tc.Expected {
			resetTestTables(t, db)
			err := db.Insert(
				&tc.Expected[i],
				&docs.DocAccountData{
					DocID:         tc.Expected[i].ID,
					AccountNumber: tc.AccountNumber,
					PeriodFrom:    tc.Date.Add(-48 * time.Hour),
					PeriodTo:      tc.Date.Add(+48 * time.Hour),
				},
			)
			if err != nil {
				t.Fatal(err)
			}

			dl, err := srv.FindVouchers("", tc.AccountNumber, tc.Date)
			if err != nil {
				fmt.Errorf("%v: %v", k, err)
			}

			if expect.ID != dl[i].ID ||
				expect.Name != dl[i].Name {
				fmt.Errorf("%v: Expect %v was %v", k, expect, dl[i])
			}
		}
	}
}

func Test_FindAccountingTxsWithoutVouchers(t *testing.T) {
	db := initGorpConn(t)

	cases := []struct {
		TxsFile       string
		LinesToImport int
		Expected      []accountingData.AccountingData
	}{
		{
			TxsFile:       path.Join(TestData, "export.csv"),
			LinesToImport: 2,
			Expected: []accountingData.AccountingData{
				{
					DocDate:          time.Date(2013, 12, 1, 0, 0, 0, 0, time.UTC),
					DateOfEntry:      time.Date(2013, 12, 2, 0, 0, 0, 0, time.UTC),
					DocNumberRange:   "B",
					DocNumber:        "7",
					PostingText:      "Gewerbeanmeldung Stadt KA",
					AmountPosted:     26.00,
					DebitAccount:     4390,
					CreditAccount:    1310,
					TaxCode:          0,
					CostUnit1:        "001",
					CostUnit2:        "",
					AmountPostedEuro: 26.00,
					Currency:         "EUR",
				},
			},
		},
	}

	for i, tc := range cases {
		docs.AddTables(db)

		err := db.DropTablesIfExists()
		if err != nil {
			t.Fatal(err)
		}
		err = db.CreateTablesIfNotExists()
		if err != nil {
			t.Fatal(err)
		}

		fh, err := os.Open(tc.TxsFile)
		if err != nil {
			t.Fatal(err)
		}

		txs := accountingTxsFileReader.NewReader(fh)
		for x := 0; x < tc.LinesToImport; x++ {
			tx, err := txs.Read()
			if err != nil {
				t.Fatal(err)
			}
			createVoucherInDB(t, db, fmt.Sprintf("test-%v_%v.pdf", i, x), tx)
		}

		a, err := FindAccountingTxsWithoutVouchers(db, path.Join(TestData, "export.csv"))
		if err != nil {
			t.Fatal("Case %v: %v", i, err)
		}

		if len(a) != len(tc.Expected) {
			t.Fatalf("Case %v: Expect len %v was %v", i, len(tc.Expected), len(a))
		}

		for i, expect := range tc.Expected {
			assert.Equal(t, expect, a[i])
		}
	}

}

func createVoucherInDB(t *testing.T, db *gorp.DbMap, filename string, tx accountingData.AccountingData) docs.Doc {
	// Erzeuge Belegeintrag desen Belegzeitraum zu einer Buchung passt.
	doc := docs.Doc{
		Name: filename,
	}

	err := db.Insert(&doc)
	if err != nil {
		t.Fatal(err)
	}

	docAccountData := docs.DocAccountData{
		DocID:         doc.ID,
		PeriodFrom:    tx.DocDate.Add(-24 * time.Hour),
		PeriodTo:      tx.DocDate.Add(24 * time.Hour),
		AccountNumber: tx.CreditAccount,
	}

	err = db.Insert(&docAccountData)
	if err != nil {
		t.Fatal(err)
	}

	docNumber := tx.DocNumberRange + tx.DocNumber
	dn := docs.DocNumber{
		DocID:  doc.ID,
		Number: docNumber,
	}
	err = db.Insert(&dn)
	if err != nil {
		t.Fatal(err)
	}

	return doc
}

func setenvTest() {
	os.Clearenv()

	os.Setenv("MYSQL_USER", "tochti")
	os.Setenv("MYSQL_PASSWORD", "123")
	os.Setenv("MYSQL_HOST", "127.0.0.1")
	os.Setenv("MYSQL_PORT", "3306")
	os.Setenv("MYSQL_DB_NAME", TestDBName)
}

func initGorpConn(t *testing.T) *gorp.DbMap {
	setenvTest()
	//gumspecs.AppName = "test"
	mysql := gumspecs.ReadMySQL()
	if mysql == nil {
		t.Fatal("Error in MySQL config")
	}

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

func setupTestDB(t *testing.T) *gorp.DbMap {
	db := initGorpConn(t)
	docs.AddTables(db)

	err := db.DropTablesIfExists()
	if err != nil {
		t.Fatal(err)
	}
	err = db.CreateTablesIfNotExists()
	if err != nil {
		t.Fatal(err)
	}

	return db
}

func resetTestTables(t *testing.T, db *gorp.DbMap) {
	docs.AddTables(db)

	err := db.DropTablesIfExists()
	if err != nil {
		t.Fatal(err)
	}
	err = db.CreateTablesIfNotExists()
	if err != nil {
		t.Fatal(err)
	}
}
