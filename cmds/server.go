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
	"github.com/tochti/gin-gum/gumspecs"
)

func main() {

	app := "accountant"

	logger := log.New(os.Stdout, fmt.Sprintf("%v: ", app), log.LstdFlags)

	gumspecs.AppName = app
	httpSpecs := gumspecs.ReadHTTPServer()
	if httpSpecs == nil {
		logger.Fatal("Error in server config")
	}
	specs, err := accountantService.ReadSpecs(app)
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
	mysql := gumspecs.ReadMySQL()
	if mysql == nil {
		logger.Fatal("Error in MySQL config")
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
