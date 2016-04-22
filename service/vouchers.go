package accountantService

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	"gopkg.in/gorp.v1"

	"github.com/tochti/docMa-accountant/service/accountingTxsFileReader"
	"github.com/tochti/docMa-handler/accountingData"
	"github.com/tochti/docMa-handler/docs"
)

func findVouchersByID(db *gorp.DbMap, id string) ([]docs.Doc, error) {
	q := fmt.Sprintf(`
		SELECT * 
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
		SELECT *
		FROM %v as docs, %v as account_data
		WHERE account_data.period_from <= ?
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
func FindAccountingTxsWithoutVouchers(db *gorp.DbMap, filePath string, voucherFilesDir string) ([]accountingData.AccountingData, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return []accountingData.AccountingData{}, err
	}

	txFile := accountingTxsFileReader.NewReader(f)

	q, d, err := makeSQLQueryFindVouchers(txFile)
	if err != nil {
		return []accountingData.AccountingData{}, err
	}

	fmt.Println(q)
	result := &[]struct{ Name string }{}
	_, err = db.Select(result, q, d...)
	if err != nil {
		return []accountingData.AccountingData{}, err
	}

	fmt.Println(result)

	return []accountingData.AccountingData{}, nil

}

func makeSQLQueryFindVouchers(txFile *accountingTxsFileReader.AccountingTxsFileReader) (string, []interface{}, error) {
	vTable1WhereTpl := fmt.Sprintf("doc_numbers.number=?")
	vTable2WhereTpl := fmt.Sprintf(`
		account_data.period_from <= ?
		AND
		account_data.period_to >= ?
		AND
		(
			 account_data.account_number = ?
			 OR
			 account_data.account_number = ?
		 )
	`)
	where1 := bytes.NewBufferString("")
	where2 := bytes.NewBufferString("")

	queryData := []interface{}{}

	// Erzeuge den ersten Eintrag da hier kein OR verwendet werden kann
	a, err := txFile.Read()
	if err != nil {
		if err == io.EOF {
			return "", []interface{}{}, nil
		}
		return "", []interface{}{}, err
	}

	where1.WriteString(vTable1WhereTpl)
	queryData = append(queryData, a.DocNumberRange+a.DocNumber)

	where2.WriteString(vTable2WhereTpl)
	queryData = append(queryData, a.DateOfEntry, a.DateOfEntry, a.CreditAccount, a.DebitAccount)

	// Erzeuge restlichen where-Abfragen anlegen
	for {
		a, err := txFile.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", []interface{}{}, err
		}

		where1.WriteString(fmt.Sprintf(" OR %v", vTable1WhereTpl))
		queryData = append(queryData, a.DocNumberRange+a.DocNumber)

		where2.WriteString(fmt.Sprintf(" OR(%v)\n", vTable2WhereTpl))
		queryData = append(queryData, a.DateOfEntry, a.DateOfEntry, a.CreditAccount, a.DebitAccount)
	}

	vTable1 := fmt.Sprintf(`
		SELECT doc_numbers.doc_id as id
		FROM %v as doc_numbers
		WHERE (%v) 
	`, docs.DocNumbersTable, where1.String())

	vTable2 := fmt.Sprintf(`
		SELECT account_data.doc_id as id
		FROM %v as account_data
		WHERE %v
	`, docs.DocAccountDataTable, where2.String())

	q := fmt.Sprintf(`
		SELECT docs.name 
		FROM %v as docs, ((%v) UNION (%v)) as ids
		WHERE docs.id = ids.id
		GROUP BY docs.name
	`, docs.DocsTable, vTable1, vTable2)

	return q, queryData, nil
}
