package driver

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/go-sql-driver/mysql"
)

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

	if host != "127.0.0.1" {

	}

	config := mysql.NewConfig()
	config.User = acct
	config.Passwd = pw
	config.Net = "tcp"
	config.Addr = host
	config.DBName = name

	// dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s", acct, pw, host, name)
	dsn := config.FormatDSN()
	dsn = dsn + "?parseTime=true&tls=false"
	os.Stderr.WriteString(dsn)
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
