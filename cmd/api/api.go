package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/torenware/go-stripe/internal/driver"
	"github.com/torenware/go-stripe/internal/models"
	mail "github.com/xhit/go-simple-mail/v2"
)

const version = "1.0.0"

type config struct {
	port int
	env  string // development | production
	db   struct {
		dsn string
	}
	stripe struct {
		secret string
		key    string
	}
	secretkey string
	frontend  string
}

// receiver type
type application struct {
	config     config
	infoLog    *log.Logger
	errorLog   *log.Logger
	version    string
	DB         *models.DBModel
	mailServer *mail.SMTPServer
}

func (app *application) serve() error {
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", app.config.port),
		Handler:           app.routes(), // TBI
		IdleTimeout:       30 * time.Second,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	app.infoLog.Printf("Starting the backend server in %s mode on port %d", app.config.env, app.config.port)

	return srv.ListenAndServe()

}

func main() {
	var config config
	//var dsn string

	flag.IntVar(&config.port, "port", 4001, "Port number")
	flag.StringVar(&config.env, "env", "development", "development|production")
	flag.Parse()

	// https://preslav.me/2020/11/10/use-dotenv-files-when-developing-your-golang-apps/
	godotenv.Load(".env.local")

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	config.stripe.key = os.Getenv("STRIPE_KEY")
	config.stripe.secret = os.Getenv("STRIPE_SECRET")

	// crypto keys
	config.secretkey = os.Getenv("SECRET_KEY")
	if config.secretkey == "" {
		errorLog.Fatalln("SECRET_KEY must be in environment")
	}
	config.frontend = os.Getenv(("FRONT_END"))
	if config.frontend == "" {
		errorLog.Fatalln("FRONT_END must be in environment")
	}

	var err error
	dsn, err := driver.ConstructDSN()
	if err != nil {
		errorLog.Println(err)
		return
	}
	config.db.dsn = dsn

	conn, err := driver.OpenDB(config.db.dsn)
	if err != nil {
		errorLog.Fatalln(err)
	}
	infoLog.Println("Database is UP")
	defer conn.Close()

	server, err := initMailserver()
	if err != nil {
		errorLog.Println(err)
		return
	}

	app := &application{
		config:     config,
		infoLog:    infoLog,
		errorLog:   errorLog,
		version:    version,
		DB:         &models.DBModel{DB: conn},
		mailServer: server,
	}

	err = app.serve()
	if err != nil {
		app.errorLog.Println(err)
		log.Fatal()
	}

}
