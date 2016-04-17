package accountantService

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tochti/docMa-handler/accountingData"
	"github.com/tochti/docMa-handler/docs"
)

type (
	AccountantService interface {
		ReadAccountingTxs() ([]accountingData.AccountingData, Error)
		FindVouchers(id string, accountNumber int, voucherDate time.Time) ([]docs.Doc, Error)
		Verify() ([]CorruptAccountingTx, Error)
	}

	Service struct {
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
func (s *Service) ReadAccountingTxs() Error {
	return nil
}

func GinReadAccountingTxsDecoder(fn AccountantService) gin.HandlerFunc {
	return func(c *gin.Context) {
		data, err := fn.ReadAccountingTxs()
		if err != nil {
			c.JSON(http.StatusBadRequest,
				ErrorResponse{
					ID:      err.ID(),
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
func (s *Service) FindVouchers(id string, accountNumber int, voucherDate time.Time) ([]docs.Doc, Error) {
	return []docs.Doc{}, nil
}

func GinFindVouchersDecoder(fn AccountantService) gin.HandlerFunc {
	return func(c *gin.Context) {
		/*
			docs, err := fn.FindVouchers(id, accountNumber, voucherDate)
			if err != nil {
				c.JSON(http.StatusBadRequest,
					ErrorResponse{
						ID:      err.ID(),
						Message: err.Err(),
					},
				)
			} else {
				c.JSON(http.StatusOK, docs)
			}
		*/
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
