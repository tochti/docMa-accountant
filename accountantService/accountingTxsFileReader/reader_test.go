package accountingTxsFileReader

import (
	"encoding/csv"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tochti/docMa-handler/accountingData"
)

var (
	TestData          = "./test_data"
	AccountingTxsFile = "./test_data/export.csv"
)

func Test_ParseFloatComma(t *testing.T) {
	assert := assert.New(t)

	f, err := ParseFloatComma("1.4440,14")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(14440.14, f)

	f, err = ParseFloatComma("1")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(1.00, f)

	f, err = ParseFloatComma("1.000")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(1000.00, f)
}

func Test_ParseGermanDate(t *testing.T) {
	tmp := "29.08.2013"
	d, err := ParseGermanDate(tmp, ".")
	if err != nil {
		t.Fatal(err)
	}

	e := time.Date(2013, time.August, 29, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, e, d)
}

func Test_ParseAccountingData(t *testing.T) {
	f, err := os.Open(path.Join(TestData, "export_ParseAccountingData.csv"))
	if err != nil {
		t.Fatal(err)
	}

	reader := csv.NewReader(f)
	reader.Comma = ';'
	reader.FieldsPerRecord = 13

	record, err := reader.Read()
	if err != nil {
		t.Fatal(err)
	}

	eAccTx := accountingData.AccountingData{
		DocDate:          time.Date(2013, time.August, 29, 0, 0, 0, 0, time.UTC),
		DateOfEntry:      time.Date(2013, time.September, 1, 0, 0, 0, 0, time.UTC),
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

	a, err := ParseAccountingData(record)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, eAccTx, a)
}

func Test_ParseAccountingData_TxPart(t *testing.T) {
	f, err := os.Open(path.Join(TestData, "export_ParseAccountingData_TxPart.csv"))
	if err != nil {
		t.Fatal(err)
	}

	reader := csv.NewReader(f)
	reader.Comma = ';'
	reader.FieldsPerRecord = 13

	record, err := reader.Read()
	if err != nil {
		t.Fatal(err)
	}

	eAccTx := accountingData.AccountingData{
		DocDate:          ZeroDate(),
		DateOfEntry:      ZeroDate(),
		DocNumberRange:   "",
		DocNumber:        "",
		PostingText:      "Strom",
		AmountPosted:     35.70,
		DebitAccount:     4240,
		CreditAccount:    0,
		TaxCode:          9,
		CostUnit1:        "310",
		CostUnit2:        "",
		AmountPostedEuro: 35.70,
		Currency:         "EUR",
	}

	a, err := ParseAccountingData(record)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, eAccTx, a)
}

func Test_IsTxPart(t *testing.T) {
	a := accountingData.AccountingData{
		DocDate:          ZeroDate(),
		DateOfEntry:      ZeroDate(),
		DocNumberRange:   "",
		DocNumber:        "",
		PostingText:      "Strom",
		AmountPosted:     35.70,
		DebitAccount:     4240,
		CreditAccount:    0,
		TaxCode:          9,
		CostUnit1:        "310",
		CostUnit2:        "",
		AmountPostedEuro: 35.70,
		Currency:         "EUR",
	}

	assert.Equal(t, true, IsTxPart(a))
}

func Test_IsNotTxPart(t *testing.T) {
	a := accountingData.AccountingData{
		DocDate:          time.Date(2013, time.August, 29, 0, 0, 0, 0, time.UTC),
		DateOfEntry:      time.Date(2013, time.September, 1, 0, 0, 0, 0, time.UTC),
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

	assert.Equal(t, false, IsTxPart(a))
}

func Test_ReadAccountingTxsFile(t *testing.T) {
	f, err := os.Open(path.Join(TestData, "export_ReadAccountingTxsFile.csv"))
	if err != nil {
		t.Fatal(err)
	}

	r := NewReader(f)

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

	tx, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, eAccTx, tx)

	eAccTx = accountingData.AccountingData{
		DocDate:          time.Date(2013, 9, 2, 0, 0, 0, 0, time.UTC),
		DateOfEntry:      time.Date(2013, 9, 10, 0, 0, 0, 0, time.UTC),
		DocNumberRange:   "",
		DocNumber:        "14",
		PostingText:      "Kopierer Oki MFP MC332",
		AmountPosted:     275.07,
		DebitAccount:     480,
		CreditAccount:    499,
		TaxCode:          0,
		CostUnit1:        "",
		CostUnit2:        "",
		AmountPostedEuro: 275.07,
		Currency:         "EUR",
	}

	tx, err = r.Read()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, eAccTx, tx)

}

func Test_ReadAccountingTxsFile_EOF(t *testing.T) {
}

func Test_ReadAccountingTxsFile_WrongFormat(t *testing.T) {
}
