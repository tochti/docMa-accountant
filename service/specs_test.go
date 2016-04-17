package accountantService

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ReadSpecs(t *testing.T) {
	os.Clearenv()

	os.Setenv("TEST_PUBLIC_DIR", "/p")
	os.Setenv("TEST_PDFVIEWER_PUBLIC_DIR", "/pv")
	os.Setenv("TEST_VOUCHERS_DIR", "/v")
	os.Setenv("TEST_ACCOUNTING_TXS_FILE", "/atf")

	s, err := ReadSpecs("test")
	if err != nil {
		t.Fatal(err)
	}

	assert := assert.New(t)
	assert.Equal(s.PublicDir, "/p")
	assert.Equal(s.PDFViewerPublicDir, "/pv")
	assert.Equal(s.VouchersDir, "/v")
	assert.Equal(s.AccountingTxsFile, "/atf")
}
