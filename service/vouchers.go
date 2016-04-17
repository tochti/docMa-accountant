package accountant

import (
	"os"

	"github.com/tochti/docMa-accountant/service/accountingTxsFileReader"
	"github.com/tochti/docMa-handler/accountingData"
)

// Find alle Buchungen die keinen Belge-Datei haben
func FindAccountingTxsWithoutVouchers(filePath string, voucherFilesDir string) ([]accountingData.AccountingData, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return []accountingData.AccountingData{}, err
	}

	txFile, err := accountingTxsFileReader.NewReader(f)
	if err != nil {
		return []accountingData.AccountingData{}, err
	}

	q, err := makeSQLQuery(txFile)

}
