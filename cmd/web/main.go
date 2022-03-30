package main

import (
	"embed"
	"encoding/gob"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/joho/godotenv"
	"github.com/torenware/go-stripe/internal/driver"
	"github.com/torenware/go-stripe/internal/models"

	vueglue "github.com/torenware/vite-go"
)

const version = "1.0.0"
const cssVersion = "1" // used for versioning assets

//go:embed "dist"
var dist embed.FS

var session *scs.SessionManager

type config struct {
	port int
	env  string // development | production
	api  string // base URI
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
	config        config
	infoLog       *log.Logger
	errorLog      *log.Logger
	templateCache map[string]*template.Template
	version       string
	DB            models.DBModel
	Session       *scs.SessionManager
	vueglue       *vueglue.VueGlue
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

	app.infoLog.Printf("Starting server in %s mode on port %d", app.config.env, app.config.port)

	return srv.ListenAndServe()

}

func main() {
	// Allow us to pass our Data map used for templates into our session.
	gob.Register(TransactionData{})
	gob.Register(templateData{})

	var config config

	flag.IntVar(&config.port, "port", 4000, "Port number")
	flag.StringVar(&config.env, "env", "development", "development|production")
	flag.StringVar(&config.api, "api", "http://localhost:4001", "Base API URI")
	// flag.StringVar(&config.db.dsn, "dsn", "", "MySQL DSN")

	flag.Parse()

	// https://preslav.me/2020/11/10/use-dotenv-files-when-developing-your-golang-apps/
	_ = godotenv.Load(".env.local")

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	config.stripe.key = os.Getenv("STRIPE_KEY")
	config.stripe.secret = os.Getenv("STRIPE_SECRET")
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


	// crypto keys
	config.secretkey = os.Getenv("SECRET_KEY")
	if config.secretkey == "" {
		errorLog.Fatalln("SECRET_KEY must be in environment")
	}
	config.frontend = os.Getenv(("FRONT_END"))
	if config.frontend == "" {
		errorLog.Fatalln("FRONT_END must be in environment")
	}

	// Initialize a new session manager and configure the session lifetime.
	session = scs.New()
	session.Store = mysqlstore.New(conn)
	session.Lifetime = 24 * time.Hour
	tc := make(map[string]*template.Template)

	app := &application{
		config:        config,
		infoLog:       infoLog,
		errorLog:      errorLog,
		templateCache: tc,
		version:       version,
		DB:            models.DBModel{DB: conn},
		Session:       session,
	}

    // set up the Vue loader
    glue, err := vueglue.NewVueGlue(dist, "dist")
    app.vueglue = glue

    err = app.serve()
	if err != nil {
		app.errorLog.Println(err)
		log.Fatal()
	}

}
