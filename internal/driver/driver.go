package driver

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/go-sql-driver/mysql"
)

func ParseDSN(dsn string) (*mysql.Config, error) {
	config, err := mysql.ParseDSN(dsn)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func ConstructDSN() (string, error) {

	var host, name, acct, pw string
	var ok bool

	if host, ok = os.LookupEnv("DB_HOST"); !ok {
		return "", errors.New("DB_HOST required")
	}

	if name, ok = os.LookupEnv("DB_NAME"); !ok {
		return "", errors.New("DB_NAME required")
	}

	if acct, ok = os.LookupEnv("DB_ACCT"); !ok {
		return "", errors.New("DB_ACCT required")
	}

	if pw, ok = os.LookupEnv("DB_PW"); !ok {
		return "", errors.New("DB_PW required")
	}

	config := mysql.NewConfig()
	config.User = acct
	config.Passwd = pw
	config.Net = "tcp"
	config.Addr = host
	config.DBName = name

	dsn := config.FormatDSN()
	dsn = dsn + "?parseTime=true&tls=false"
	return dsn, nil
}

func OpenDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		os.Stderr.WriteString("open failed")
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		os.Stderr.WriteString("ping of db failed")
		fmt.Println(err)
		return nil, err
	}

	return db, nil
}
