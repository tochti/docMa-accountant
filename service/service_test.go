package accountantService

import (
	"os"
	"testing"
	"time"

	"github.com/tochti/docMa-handler/accountingData"
)

func Test_ReadAccountingTxs(t *testing.T) {
	cases := map[string]struct {
		TxFile   string
		Expected accountingData.AccountingData
	}{
		"default": {
			TxFile: "./test_data/readAccountingTxs.csv",
			Expected: accountingData.AccountingData{
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
	}

	db := setupTestDB(t)
	for k, tc := range cases {
		os.Setenv("TEST_ACCOUNTING_TXS_FILE", tc.TxFile)
		s, err := ReadSpecs("TEST")
		if err != nil {
			t.Errorf("%v: %v", k, err)
		}

		srv := Service{
			DB:    db,
			Specs: &s,
		}

		al, err := srv.ReadAccountingTxs()
		if err != nil {
			t.Errorf("%v: %v", k, err)
		}

		if tc.Expected.DocNumber != al[0].DocNumber {
			t.Errorf("%v: Expect %v was %v", k, tc.Expected, al[0])
		}
	}
}
