package cmds

import (
	"path"

	"github.com/gin-gonic/gin"
	"github.com/tochti/gin-gum/gumspecs"
)

func main() {

	httpSpecs := gumspecs.ReadHTTPServer()
	specs := service.ReadSepcs()

	service := &accountantService.Service{
		Logger: log,
		Specs:  specs,
	}

	htmlDir := path.Join(specs.PublicDir, "html")

	router := gin.New()
	router.Static("/", htmlDir)
	router.Static("/public", specs.PublicDir)
	router.Static("/data", specs.VouchersDir)
	router.Static("/pdfviewer", specs.PDFViewerPublic)

	v1 := router.Group("/v1")
	v1.GET("/accounting_txs",
		accountantService.GinAccountingTxsDecoder(service.AccountingTxs))
	v1.GET("/vouchers",
		accountantService.GinFindVouchersDecoder(service.AccountingTxs))
	v1.GET("/verify",
		accountantService.GinVerifyDecoder(service.AccountingTxs))

	router.Run(httpSpecs.String())
}
