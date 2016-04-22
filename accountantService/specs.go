package accountantService

import "github.com/kelseyhightower/envconfig"

type (
	Specs struct {
		PublicDir          string `envconfig:"PUBLIC_DIR"`
		PDFViewerPublicDir string `envconfig:"PDFVIEWER_PUBLIC_DIR"`
		VouchersDir        string `envconfig:"VOUCHERS_DIR"`
		AccountingTxsFile  string `envconfig:"ACCOUNTING_TXS_FILE"`
	}
)

// Lese Konfigurations Parameter aus Umgebungsvariablen
func ReadSpecs(prefix string) (Specs, error) {
	s := Specs{}
	err := envconfig.Process(prefix, &s)
	if err != nil {
		return Specs{}, err
	}

	return s, nil

}
