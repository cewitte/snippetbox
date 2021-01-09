package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/cewitte/snippetbox/pkg/models/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golangcollege/sessions"
)

// Define an application struct to hold the application-wide dependencies for the web application. For now we'll only include fields for the two custom loggers, but we'll add more to it as the build progresses.
type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	session       *sessions.Session             // Add a new session field to the application struct
	snippets      *mysql.SnippetModel           // Add a snippets field to the application struct. This will allow us to make the SnippetModel object available to our handlers.
	templateCache map[string]*template.Template // Add a templateCache field to the application struct.
}

func main() {
	// Define a new command-line flag with the name "addr", a default value of ":4000" and some short help text explaining what the flag controls. The value of the flag will be stored in the addr variable at runtime
	addr := flag.String("addr", ":4000", "HTTP network address")

	// Define a new command-line flag for the MySQL DSN string.
	dsn := flag.String("dsn", "web:1234@/snippetbox?parseTime=true", "MySQL data source name")

	// Define a new command-line flag for the session secret (a random key which will be used to encrypt and authenticate session cookies). It should be 32 bytes long.
	secret := flag.String("secret", "s6Ndh+pPbnzHbS*+9Pk8qGWhTzbpa@ge", "Secret key")

	// Importantly, we use the flag.Parse() function to parse the command line flag.
	// This reads the command-line flag value and assigns it to the addr variable. You need to call this *before* you use the addr variable otherwise it will always contain the default value of ":4000". If any errors are encountered during parsing the application will be terminated.
	flag.Parse()

	// Use log.New() to create a logger for writing information messages. This takes 3 parameters: the destination to write logs to (os.Stdout), a string prefix for message (INFO followed by a tab), and flags to indicate what additional information to incluce (local date and time). Note that the flags are joined using the bitwise OR operator |.
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)

	// Create a logger for writing error messages in the same way, but use stderr as the destination and use the log.Lshortfile flag to include the relevant filename and line number.
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// To keep the main() function tidy I've put the code for creating a connection cool into the separate openDB() funcion below. We pass openDB() the DSN from the command-line flag.
	db, err := openDB(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}

	// Added here to inform a successful connection with the database
	infoLog.Println("DB connection successful")

	// We also defer a call to db.Close(), so that the connection pool is closed before the main() function exits.
	defer db.Close()

	// Initialize a new template cache...
	templateCache, err := newTemplateCache("./ui/html/")
	if err != nil {
		errorLog.Fatal(err)
	}

	// Use the sessions.New() function to initialize a new session manager, passing in the secret key as a parameter. Then we configure it so sessions always expire after 12 hours.
	session := sessions.New([]byte(*secret))
	session.Lifetime = 12 * time.Hour

	// Initialize a new instance of application containing the dependencies.
	app := &application{
		errorLog:      errorLog,
		infoLog:       infoLog,
		session:       session,
		snippets:      &mysql.SnippetModel{DB: db}, // Initialize mysql.SnippetModel instance and add it to the application dependencies.
		templateCache: templateCache,
	}

	// Initialize a tls.Config struct to hold the non-default TLS settings we want the server to use.
	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,
		CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	// Swap the route declarations to use the application struct's methods as the handler function
	mux := http.NewServeMux()
	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/snippet", app.showSnippet)
	mux.HandleFunc("snipppet/create", app.createSnippet)

	// Initialize a new http.Server struct. We set the Addr and Handler fields so that the server uses the same network address and routes as before and set the ErrorLog field so that the server now uses the customer ErrorLog logger in the event of any problems.
	// Set the server's TLSConfig field to use the tlsConfig variable just created.
	srv := &http.Server{
		Addr:      *addr,
		ErrorLog:  errorLog,
		Handler:   app.routes(), // Call the new app.routes() method
		TLSConfig: tlsConfig,
		// Add Idle, Read and Write timeouts to the server.
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// The value returned from the flag.String() function is a pointer to the flag value, not the value itself. So we need to dereference the pointer (i.e. prefix it with the * symbol) before using it. Note that we're using the log.Printf() function to interpolate the address with the log message.
	// Write messages using the 2 new loggers, instead of the standard logger.
	infoLog.Printf("Starting server on %s", *addr)
	// Use the ListenAndServeTLS() method to start the HTTPS server. We pass in the paths to the TLS certificate and corresponding private keys as the two parameters.
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	errorLog.Fatal(err)
}

// The openDB() function wraps sql.Open() and returns a sql.DB connection pool for a given DSN
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
