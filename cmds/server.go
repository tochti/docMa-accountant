package main

import (
	"fmt"
	"log"
	"os"
	"path"

	"gopkg.in/gorp.v1"

	"github.com/gin-gonic/gin"
	"github.com/tochti/docMa-accountant/accountantService"
	"github.com/tochti/docMa-handler"
	"github.com/tochti/docMa-handler/docs"
)

const (
	APPNAME = "accountant"
)

func main() {

	logger := log.New(os.Stdout, fmt.Sprintf("%v: ", APPNAME), log.LstdFlags)

	httpSpecs, err := accountantService.ReadHTTPServerSpecs(APPNAME)
	if err != nil {
		logger.Fatal(err)
	}
	specs, err := accountantService.ReadSpecs(APPNAME)
	if err != nil {
		logger.Fatal(err)
	}

	db := initDB(logger)
	service := &accountantService.Service{
		DB:    db,
		Log:   logger,
		Specs: &specs,
	}

	htmlDir := path.Join(specs.PublicDir, "html")

	router := gin.New()
	router.Use(bebber.Serve("/", bebber.LocalFile(htmlDir, false)))
	router.Static("/html", htmlDir)
	router.Static("/public", specs.PublicDir)
	router.Static("/data", specs.VouchersDir)
	router.Static("/pdfviewer", specs.PDFViewerPublicDir)

	v1 := router.Group("/v1")
	v1.GET("/accounting_txs",
		accountantService.GinReadAccountingTxsDecoder(service))
	v1.GET("/vouchers",
		accountantService.GinFindVouchersDecoder(service))
	v1.GET("/verify",
		accountantService.GinVerifyDecoder(service))

	router.Run(httpSpecs.String())
}

func initDB(logger *log.Logger) *gorp.DbMap {
	mysql, err := accountantService.ReadMySQLSpecs(APPNAME)
	if err != nil {
		logger.Fatal(err)
	}

	c, err := mysql.DB()
	if err != nil {
		logger.Fatal(err)
	}

	db := &gorp.DbMap{
		Db: c,
		Dialect: gorp.MySQLDialect{
			"InnonDB",
			"UTF8",
		},
	}

	docs.AddTables(db)

	err = db.DropTablesIfExists()
	if err != nil {
		logger.Fatal(err)
	}
	err = db.CreateTablesIfNotExists()
	if err != nil {
		logger.Fatal(err)
	}

	return db

}
