package accountantService

import (
	"database/sql"
	"fmt"
	"log"
	"testing"

	"github.com/tochti/docMa-handler/docs"

	"gopkg.in/gorp.v1"
)

var (
	DBCONNECTION = &sql.DB{}
)

func init() {
	mysql, err := ReadMySQLSpecs("test")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(mysql.String())

	DBCONNECTION, err = mysql.DB()
	if err != nil {
		log.Fatal(err)
	}
}

func initGorpConn(t *testing.T) *gorp.DbMap {
	return &gorp.DbMap{
		Db: DBCONNECTION,
		Dialect: gorp.MySQLDialect{
			"InnonDB",
			"UTF8",
		},
	}
}

func SetupTestDB(t *testing.T) *gorp.DbMap {
	db := initGorpConn(t)

	docs.AddTables(db)

	err := db.DropTablesIfExists()
	if err != nil {
		t.Fatal(err)
	}
	err = db.CreateTablesIfNotExists()
	if err != nil {
		t.Fatal(err)
	}

	return db
}
