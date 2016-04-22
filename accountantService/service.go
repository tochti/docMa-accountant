package accountantService

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"gopkg.in/gorp.v1"

	"github.com/gin-gonic/gin"
	"github.com/tochti/docMa-accountant/accountantService/accountingTxsFileReader"
	"github.com/tochti/docMa-handler/accountingData"
	"github.com/tochti/docMa-handler/docs"
)

type (
	AccountantService interface {
		ReadAccountingTxs() ([]accountingData.AccountingData, error)
		FindVouchers(id string, accountNumber int, voucherDate time.Time) ([]docs.Doc, error)
		Verify() ([]CorruptAccountingTx, Error)
	}

	Service struct {
		DB    *gorp.DbMap
		Log   *log.Logger
		Specs *Specs
	}

	CorruptAccountingTx struct {
		AccountingData accountingData.AccountingData
		LineNumber     int
		Message        string
	}

	ErrorResponse struct {
		ID      int64
		Message error
	}

	Error interface {
		ID() int64
		Error() string
	}

	fail struct {
		id  int64
		err error
	}
)

func NewError(id int64, err error) fail {
	return fail{
		id:  id,
		err: err,
	}
}

func (f fail) ID() int64 {
	return f.id
}

func (f fail) Error() string {
	return f.err.Error()
}

// Lese CSV Datei mit Buchungen und weise diesen die passenden Buchungsbelege zu.
func (s *Service) ReadAccountingTxs() ([]accountingData.AccountingData, error) {
	f, err := os.Open(s.Specs.AccountingTxsFile)
	if err != nil {
		return []accountingData.AccountingData{}, err
	}

	reader := accountingTxsFileReader.NewReader(f)

	al := []accountingData.AccountingData{}
	for {
		tx, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}

			return []accountingData.AccountingData{}, err
		}
		al = append(al, tx)

	}

	return al, nil

}

func GinReadAccountingTxsDecoder(fn AccountantService) gin.HandlerFunc {
	return func(c *gin.Context) {
		data, err := fn.ReadAccountingTxs()
		if err != nil {
			c.JSON(http.StatusBadRequest,
				ErrorResponse{
					ID:      1,
					Message: err,
				},
			)
		} else {
			c.JSON(http.StatusOK, data)
		}
	}
}

// Find alle Buchungsbelege die entweder
// die korrekte Buchungsbeleg-Nummer haben
// oder
// das Buchungsbelge-Datum befindes sich innerhalb eines Zeitraums
// und die Kontonummer stimmt überein
func (s *Service) FindVouchers(id string, accountNumber int, voucherDate time.Time) ([]docs.Doc, error) {
	// Ist id übergeben benutze diese um Buchungsbeleg zu finden
	if id != "" {
		return findVouchersByID(s.DB, id)
	} else if accountNumber > 0 && voucherDate.After(zeroDate()) {
		return findVouchersByAccountNumber(s.DB, accountNumber, voucherDate)
	}

	return []docs.Doc{}, fmt.Errorf("Wrong arguments")
}

func GinFindVouchersDecoder(fn AccountantService) gin.HandlerFunc {
	return func(c *gin.Context) {

		id := c.DefaultQuery("id", "")
		accountNumberTmp := c.DefaultQuery("account_number", "")
		voucherDateTmp := c.DefaultQuery("voucher_date", "")

		var accountNumber int
		if accountNumberTmp == "" {
			accountNumber = -1
		} else {
			var err error
			accountNumber, err = strconv.Atoi(accountNumberTmp)
			if err != nil {
				c.JSON(http.StatusBadRequest,
					ErrorResponse{
						ID:      2,
						Message: err,
					},
				)
				return
			}
		}

		var voucherDate time.Time
		if voucherDateTmp == "" {
			voucherDate = zeroDate()
		} else {
			var err error
			voucherDate, err = time.Parse(time.RFC3339, voucherDateTmp)
			if err != nil {
				c.JSON(http.StatusBadRequest,
					ErrorResponse{
						ID:      2,
						Message: err,
					},
				)
				return
			}
		}

		docs, err := fn.FindVouchers(id, accountNumber, voucherDate)
		if err != nil {
			c.JSON(http.StatusBadRequest,
				ErrorResponse{
					ID:      2,
					Message: err,
				},
			)
		} else {
			c.JSON(http.StatusOK, docs)
		}
	}
}

// Überprüfe Buchungsbelege auf logische Fehler.
// Welche Belege fehlen, d. h. für welche Datensätze in der CSV keine PDF existiert.
// Haben zwei PDFs die gleiche Belegnummer.
func (s *Service) Verify() ([]CorruptAccountingTx, Error) {
	// [] := FindAccountingTxsWithoutVouchers(s.Specs)
	return []CorruptAccountingTx{}, nil
}

func GinVerifyDecoder(fn AccountantService) gin.HandlerFunc {
	return func(c *gin.Context) {
		corrupts, err := fn.Verify()
		if len(corrupts) > 0 {
			c.JSON(http.StatusTeapot, corrupts)
		} else if err != nil {
			c.JSON(http.StatusBadRequest,
				ErrorResponse{
					ID:      err.ID(),
					Message: err,
				},
			)
		} else {
			c.JSON(http.StatusOK, "")
		}
	}
}

func zeroDate() time.Time {
	return time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
}
