package accountantService

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tochti/docMa-handler/accountingData"
	"github.com/tochti/docMa-handler/docs"
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

	db := SetupTestDB(t)
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

func TestService_FindVouchersByID(t *testing.T) {

	cases := map[string]struct {
		URL       string
		DocNumber string
		Doc       docs.Doc
		Expected  string
	}{
		"default": {
			URL:       "/v?id=B7",
			DocNumber: "B7",
			Doc: docs.Doc{
				ID:            1,
				Name:          "alfonso",
				DateOfScan:    time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
				DateOfReceipt: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			Expected: `[{"id":1,"name":"alfonso","barcode":"","date_of_scan":"2000-01-01T00:00:00Z","date_of_receipt":"2000-01-01T00:00:00Z","note":""}]`,
		},
	}

	for k, tc := range cases {
		db := SetupTestDB(t)
		srv := &Service{
			DB: db,
		}

		err := db.Insert(&tc.Doc, &docs.DocNumber{tc.Doc.ID, tc.DocNumber})
		if err != nil {
			t.Fatal(err)
		}

		req, _ := http.NewRequest("GET", tc.URL, bytes.NewBufferString(""))
		w := httptest.NewRecorder()

		gin.SetMode(gin.ReleaseMode)
		r := gin.New()
		r.GET("/v", GinFindVouchersDecoder(srv))

		r.ServeHTTP(w, req)

		if strings.TrimRight(w.Body.String(), "\n") != tc.Expected {
			t.Errorf("%v: Expect %v was %v", k, tc.Expected, w.Body.String())

		}
	}
}

func TestService_FindVouchersByAccountNumber(t *testing.T) {
	cases := map[string]struct {
		URL           string
		AccountNumber int
		Date          time.Time
		Doc           docs.Doc
		Expected      string
	}{
		"default": {
			URL:           "/v?account_number=1400&voucher_date=2000-01-01T00:00:00Z",
			AccountNumber: 1400,
			Date:          time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			Doc: docs.Doc{
				ID:            1,
				Name:          "alfonso",
				DateOfScan:    time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
				DateOfReceipt: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			Expected: `[{"id":1,"name":"alfonso","barcode":"","date_of_scan":"2000-01-01T00:00:00Z","date_of_receipt":"2000-01-01T00:00:00Z","note":""}]`,
		},
	}

	for k, tc := range cases {
		db := SetupTestDB(t)
		srv := &Service{
			DB: db,
		}

		err := db.Insert(
			&tc.Doc,
			&docs.DocAccountData{
				DocID:         tc.Doc.ID,
				AccountNumber: tc.AccountNumber,
				PeriodFrom:    tc.Date.Add(-48 * time.Hour),
				PeriodTo:      tc.Date.Add(+48 * time.Hour),
			},
		)
		if err != nil {
			t.Fatal(err)
		}

		req, _ := http.NewRequest("GET", tc.URL, bytes.NewBufferString(""))
		w := httptest.NewRecorder()

		gin.SetMode(gin.ReleaseMode)
		r := gin.New()
		r.GET("/v", GinFindVouchersDecoder(srv))

		r.ServeHTTP(w, req)

		if strings.TrimRight(w.Body.String(), "\n") != tc.Expected {
			t.Errorf("%v: Expect %v was %v", k, tc.Expected, w.Body.String())

		}
	}
}
