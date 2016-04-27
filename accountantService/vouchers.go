package accountantService

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"time"

	"gopkg.in/gorp.v1"

	"github.com/tochti/docMa-accountant/accountantService/accountingTxsFileReader"
	"github.com/tochti/docMa-handler/accountingData"
	"github.com/tochti/docMa-handler/docs"
)

func findVouchersByID(db *gorp.DbMap, id string) ([]docs.Doc, error) {
	q := fmt.Sprintf(`
		SELECT docs.* 
		FROM %v as docs, %v as doc_numbers 
		WHERE doc_numbers.number=?
		AND docs.id=doc_numbers.doc_id
		`, docs.DocsTable, docs.DocNumbersTable)

	dl := []docs.Doc{}
	_, err := db.Select(&dl, q, id)
	if err != nil {
		return []docs.Doc{}, err
	}

	return dl, nil

}

func findVouchersByAccountNumber(db *gorp.DbMap, accountNumber int, date time.Time) ([]docs.Doc, error) {
	q := fmt.Sprintf(`
		SELECT docs.*
		FROM %v as docs, %v as account_data
		WHERE account_data.account_number=?
		AND account_data.period_from <= ?
		AND account_data.period_to >= ?
		AND docs.id=account_data.doc_id
	`, docs.DocsTable, docs.DocAccountDataTable)

	dl := []docs.Doc{}
	_, err := db.Select(&dl, q, accountNumber, date, date)
	if err != nil {
		return []docs.Doc{}, err
	}

	return dl, nil
}

// Find alle Buchungen die keinen Belge-Datei haben s. auch Verify Funktion
func FindAccountingTxsWithoutVouchers(db *gorp.DbMap, filePath string) ([]accountingData.AccountingData, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return []accountingData.AccountingData{}, err
	}

	txFile := accountingTxsFileReader.NewReader(f)

	voucherIDs := []string{}
	for {
		tx, err := txFile.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return []accountingData.AccountingData{}, nil
		}

		voucherIDs = append(voucherIDs, tx.DocNumberRange+tx.DocNumber)
	}

	voucherNumbers, err := findVoucherNumbersByIDs(db, voucherIDs...)
	if err != nil {
		return []accountingData.AccountingData{}, err
	}

	// Bau MapIndex auf um heraus zu finden welche Buchungen keine Buchungsnummer in der DB haben
	tmp := map[string]struct{}{}
	for _, v := range voucherNumbers {
		tmp[v.Number] = struct{}{}
	}

	err = txFile.Reset()
	if err != nil {
		return []accountingData.AccountingData{}, err
	}

	txWithNoDBEntry := []accountingData.AccountingData{}
	for {
		tx, err := txFile.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return []accountingData.AccountingData{}, err
		}
		_, ok := tmp[tx.DocNumberRange+tx.DocNumber]
		if !ok {
			txWithNoDBEntry = append(txWithNoDBEntry, tx)
		}
	}

	return txWithNoDBEntry, nil

}

func findVoucherNumbersByIDs(db *gorp.DbMap, vouchers ...string) ([]docs.DocNumber, error) {

	or := bytes.NewBufferString("")
	for range vouchers[1:] {
		or.WriteString("OR doc_numbers.number = ? ")
	}

	q := fmt.Sprintf(`
		SELECT * 
		FROM %v as doc_numbers 
		WHERE doc_numbers.number = ?
		%v
		`, docs.DocNumbersTable, or)

	voucherNumbers := []docs.DocNumber{}
	_, err := db.Select(&voucherNumbers, q, InterfaceSlice(vouchers)...)
	if err != nil {
		return []docs.DocNumber{}, err
	}

	return voucherNumbers, nil
}

func InterfaceSlice(slice interface{}) []interface{} {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		panic("InterfaceSlice() given a non-slice type")
	}

	ret := make([]interface{}, s.Len())

	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}

	return ret
}
