package accountingTxsFileReader

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/tochti/docMa-handler/accountingData"
)

/*
// Lese Buchungen aus Datei
func ReadAccountingTxFile(filename string) ([]accountData.AccountData, error) {
	f, err := os.Open(fName)
	if err != nil {
		return err
	}
	reader := csv.NewReader(f)
	reader.Comma = ';'
	reader.FieldsPerRecord = 13
	// Skip Headline
	reader.Read()
	for {
		r := AccData{}
		err := UnmarshalAccData(reader, &r)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		} else {
			if (r.Belegdatum.IsZero() == true) &&
				(r.Buchungsdatum.IsZero() == true) &&
				(r.Belegnummernkreis == "") &&
				(r.Belegnummer == "") {
				continue
			}
			*ad = append(*ad, r)
		}
	}
	return nil
}
*/

type (
	AccountingTxsFileReader struct {
		fh        *os.File
		csvReader *csv.Reader
		line      int
	}
)

func NewReader(fh *os.File) *AccountingTxsFileReader {
	reader := csv.NewReader(fh)
	reader.Comma = ';'
	reader.FieldsPerRecord = 13

	// Skip Headline
	reader.Read()

	return &AccountingTxsFileReader{
		fh:        fh,
		csvReader: reader,
		line:      1,
	}
}

func (r *AccountingTxsFileReader) Read() (accountingData.AccountingData, error) {
	r.line++

	record, err := r.csvReader.Read()
	if err != nil {
		return accountingData.AccountingData{}, err
	}

	a, err := ParseAccountingData(record)
	if err != nil {
		err = fmt.Errorf("line: %v %v", r.line, err)
		return accountingData.AccountingData{}, err
	}

	// Ist der ausgelesene Datensatz eine Teil Buchung gehe zu nächstem Datensatz
	if IsTxPart(a) {
		return r.Read()
	}

	return a, nil
}

func (r *AccountingTxsFileReader) Reset() error {
	_, err := r.fh.Seek(0, 0)
	if err != nil {
		return err
	}

	// Skip Headline
	r.Read()

	return nil

}

func NewParseError(record []string, field int, msg string) error {
	s := "record: %v, field: %v - %v"
	return fmt.Errorf(s, record, field, msg)
}

// Erstelle aus einem CSV Record, falls möglich, ein AccountingData Objekt
// todo(tochti): Zerlege Funktion
func ParseAccountingData(record []string) (accountingData.AccountingData, error) {

	acc := accountingData.AccountingData{}

	// Sind die ersten vier Felder leer ist der Eintrag ein Teil einer Buchung
	if (record[0] == "") && (record[1] == "") &&
		(record[2] == "") && (record[3] == "") {
		acc.DateOfEntry = ZeroDate()
		acc.DocDate = ZeroDate()
		acc.DocNumberRange = ""
		acc.DocNumber = ""
	} else {
		date, err := ParseGermanDate(record[0], ".")
		if err != nil {
			msg := "Cannot parse voucher date - " + err.Error()
			return accountingData.AccountingData{},
				NewParseError(record, 1, msg)
		}
		acc.DocDate = date

		date, err = ParseGermanDate(record[1], ".")
		if err != nil {
			msg := "Cannot parse date of entry - " + err.Error()
			return accountingData.AccountingData{},
				NewParseError(record, 2, msg)
		}
		acc.DateOfEntry = date

		acc.DocNumberRange = record[2]
		acc.DocNumber = record[3]
	}

	acc.PostingText = record[4]

	fl, err := ParseFloatComma(record[5])
	if err != nil {
		msg := "amount posted is not a float number - " + err.Error()
		return accountingData.AccountingData{},
			NewParseError(record, 6, msg)
	}
	acc.AmountPosted = fl

	if record[6] != "" {
		in, err := strconv.Atoi(record[6])
		if err != nil {
			msg := "debit account is not a integer - " + err.Error()
			return accountingData.AccountingData{},
				NewParseError(record, 7, msg)
		}

		acc.DebitAccount = in
	} else {
		acc.DebitAccount = 0
	}

	if record[7] != "" {
		in, err := strconv.Atoi(record[7])
		if err != nil {
			msg := "credit account is not a integer - " + err.Error()
			return accountingData.AccountingData{},
				NewParseError(record, 8, msg)
		}

		acc.CreditAccount = in
	} else {
		acc.CreditAccount = 0
	}

	if record[8] != "" {
		in, err := strconv.Atoi(record[8])
		if err != nil {
			msg := "tax code is not a integer - " + err.Error()
			return accountingData.AccountingData{},
				NewParseError(record, 9, msg)
		}
		acc.TaxCode = in
	} else {
		acc.TaxCode = 0
	}

	acc.CostUnit1 = record[9]
	acc.CostUnit2 = record[10]

	fl, err = ParseFloatComma(record[11])
	if err != nil {
		msg := "amount posted euro is not a float number - " + err.Error()
		return accountingData.AccountingData{},
			NewParseError(record, 12, msg)
	}
	acc.AmountPostedEuro = fl

	acc.Currency = record[12]

	return acc, nil

}

// Ist der übergeben Datensatz eine Teilbuchung?
func IsTxPart(a accountingData.AccountingData) bool {
	if (a.DocDate.IsZero() == true) &&
		(a.DateOfEntry.IsZero() == true) &&
		(a.DocNumberRange == "") &&
		(a.DocNumber == "") {
		return true
	}

	return false
}

func ZeroDate() time.Time {
	return time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)
}

// Erzeuge aus XX.XX.XXXX Datum ein time.Time Objekt
func ParseGermanDate(d string, sep string) (time.Time, error) {
	tmp := strings.Split(d, sep)

	dtmp, err := strconv.ParseInt(tmp[0], 10, 0)
	if err != nil {
		return ZeroDate(), err
	}
	mtmp, err := strconv.ParseInt(tmp[1], 10, 0)
	if err != nil {
		return ZeroDate(), err
	}
	ytmp, err := strconv.ParseInt(tmp[2], 10, 0)
	if err != nil {
		return ZeroDate(), err
	}
	m, err := Month(mtmp)
	if err != nil {
		return ZeroDate(), err
	}
	return time.Date(int(ytmp), m, int(dtmp), 0, 0, 0, 0, time.UTC), nil
}

// Erzeuge aus einem String wie 1.000,50 ein Float Objekt 1000.50
func ParseFloatComma(s string) (float64, error) {
	fStr := strings.Replace(s, ".", "", -1)
	fStr = strings.Replace(fStr, ",", ".", -1)

	f, err := strconv.ParseFloat(fStr, 64)
	if err != nil {
		return 0, err
	} else {
		return f, nil
	}

}

// Gebe zurück time.$Month für gegeben Zahl
func Month(m int64) (time.Month, error) {
	if (m < 1) || (m > 12) {
		return time.April, errors.New("Month out of range")
	}
	months := []time.Month{
		time.January,
		time.February,
		time.March,
		time.April,
		time.May,
		time.June,
		time.July,
		time.August,
		time.September,
		time.October,
		time.November,
		time.December,
	}

	return months[m-1], nil
}
