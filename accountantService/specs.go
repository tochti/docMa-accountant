package accountantService

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/kelseyhightower/envconfig"
)

const (
	DefaultLocation = "Europe/Berlin"
)

type (
	Specs struct {
		PublicDir          string `envconfig:"PUBLIC_DIR"`
		PDFViewerPublicDir string `envconfig:"PDFVIEWER_PUBLIC_DIR"`
		VouchersDir        string `envconfig:"VOUCHERS_DIR"`
		AccountingTxsFile  string `envconfig:"ACCOUNTING_TXS_FILE"`
	}

	MySQL struct {
		User     string `envconfig:"MYSQL_USER"`
		Password string `envconfig:"MYSQL_PASSWORD"`
		Host     string `envconfig:"MYSQL_HOST"`
		Port     int    `envconfig:"MYSQL_PORT"`
		DBName   string `envconfig:"MYSQL_DB_NAME"`
		Location string `envconfig:"MYSQL_LOCATION"`
	}

	HTTPServerSpecs struct {
		Host string `envconfig:"http_host"`
		Port int    `envconfig:"http_port"`
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

// Read MySQL settings from env
func ReadMySQLSpecs(prefix string) (MySQL, error) {
	mysql := MySQL{}

	err := envconfig.Process(prefix, &mysql)

	if err != nil {
		log.Fatal("Unable to load mysql specs", err)
	}

	if mysql.Host == "" {
		return MySQL{}, fmt.Errorf("Missing Host")
	} else if mysql.Port == 0 {
		return MySQL{}, fmt.Errorf("Missing Port")
	} else if mysql.DBName == "" {
		return MySQL{}, fmt.Errorf("Missing database name")
	} else if mysql.User == "" {
		return MySQL{}, fmt.Errorf("Missing user")
	} else if mysql.Password == "" {
		return MySQL{}, fmt.Errorf("Missing password")
	}

	return mysql, nil
}

func (m *MySQL) DB() (*sql.DB, error) {
	db, err := sql.Open("mysql", m.String())
	if err != nil {
		return db, err
	}

	return db, nil
}

func (m *MySQL) String() string {

	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		m.User,
		m.Password,
		m.Host,
		m.Port,
		m.DBName,
	)
}

func ReadHTTPServerSpecs(prefix string) (HTTPServerSpecs, error) {
	specs := HTTPServerSpecs{}
	err := envconfig.Process(prefix, &specs)
	if err != nil {
		log.Fatal(err)
	}

	if specs.Host == "" {
		return HTTPServerSpecs{}, fmt.Errorf("Missing host")
	} else if specs.Port == 0 {
		return HTTPServerSpecs{}, fmt.Errorf("Missing port")
	}

	return specs, nil
}

func (s HTTPServerSpecs) String() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}
