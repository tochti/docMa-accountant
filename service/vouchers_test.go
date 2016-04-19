package accountantService

import (
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tochti/docMa-accountant/service/accountingTxsFileReader"
	"github.com/tochti/docMa-handler/accountingData"
	"github.com/tochti/docMa-handler/docs"
	"github.com/tochti/gin-gum/gumspecs"
	"gopkg.in/gorp.v1"
)

var (
	TestData   = "./test_data"
	TestDBName = "testing"
)

func Test_FindAccountingTxsWithoutVouchers(t *testing.T) {
	txFile := path.Join(TestData, "export.csv")
	db := initGorpConn(t)
	createVouchersInDB(t, db, txFile)

	// remove db entry

	a, err := FindAccountingTxsWithoutVouchers(db, path.Join(TestData, "export.csv"), TestData)
	fmt.Println(a)
	if err != nil {
		t.Fatal(err)
	}

	// 29.08.2013;01.09.2013;"B";"7";"Gewerbeanmeldung Stadt KA";26,00;4390;1210;0;"001";"";26,00;"EUR"
	eAccTx := accountingData.AccountingData{
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
	}

	accTx := a[0]
	assert.Equal(t, eAccTx, accTx)

}

func createVouchersInDB(t *testing.T, db *gorp.DbMap, txFile string) {
	docs.AddTables(db)

	err := db.DropTablesIfExists()
	if err != nil {
		t.Fatal(err)
	}
	err = db.CreateTablesIfNotExists()
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.Open(txFile)
	if err != nil {
		t.Fatal(err)
	}
	r := accountingTxsFileReader.NewReader(f)

	// Erzeuge Belegeintrag desen Belegzeitraum zu einer Buchung passt.
	a, err := r.Read()

	doc := docs.Doc{
		Name: "test1.pdf",
	}

	err = db.Insert(&doc)
	if err != nil {
		t.Fatal(err)
	}

	docAccountData := docs.DocAccountData{
		DocID:         doc.ID,
		PeriodFrom:    a.DocDate.Add(-24 * time.Hour),
		PeriodTo:      a.DocDate.Add(24 * time.Hour),
		AccountNumber: a.CreditAccount,
	}

	err = db.Insert(&docAccountData)
	if err != nil {
		t.Fatal(err)
	}

	// Erzeuge einen Belegeintrag desen Belegnummer zu einer Buchung passt.
	a, err = r.Read()

	doc = docs.Doc{
		Name: "test2.pdf",
	}

	err = db.Insert(&doc)
	if err != nil {
		t.Fatal(err)
	}

	docNumber := a.DocNumberRange + a.DocNumber
	dn := docs.DocNumber{
		DocID:  doc.ID,
		Number: docNumber,
	}
	err = db.Insert(&dn)
	if err != nil {
		t.Fatal(err)
	}
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
