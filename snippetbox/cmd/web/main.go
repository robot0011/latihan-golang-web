// Package main is the main package of the Go program.
// It contains the entry point function `main` which is executed when the program starts.
// This package is typically used for creating executable programs.
package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"

	"snippetbox.faqih/snippetbox/internal/models"

	_ "github.com/go-sql-driver/mysql" // New import
)

type config struct {
	addr      string
	staticDir string
}

type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	snippets *models.SnippetModel
}

var cfg config

func main() {

	// addr := flag.String("addr", ":8000", "HTTP network address")
	flag.StringVar(&cfg.addr, "addr", ":8000", "HTTP network address")
	flag.StringVar(&cfg.staticDir, "static-dir", "./ui/static", "Path to static assets")
	dsn := flag.String("dsn", "web:admin1234@/snippetbox?parseTime=true", "MySQL data source name")

	flag.Parse()

	// f, err := os.OpenFile("/tmp/info.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer f.Close()

	// infoLog := log.New(f, "INFO\t", log.Ldate|log.Ltime)
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)

	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := openDB(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()
	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
		snippets: &models.SnippetModel{DB: db},
	}

	mux := http.NewServeMux()
	// Create a file server which serves files out of the "./ui/static" directory.
	// Note that the path given to the http.Dir function is relative to the project
	// directory root.
	// fileServer := http.FileServer(http.Dir("./ui/static/"))
	fileServer := http.FileServer(http.Dir(cfg.staticDir))
	// Use the mux.Handle() function to register the file server as the handler for
	// all URL paths that start with "/static/". For matching paths, we strip the
	// "/static" prefix before the request reaches the file server.
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))
	// Initialize a new instance of our application struct, containing the
	// dependencies.

	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/snippet/view", app.snippetView)
	mux.HandleFunc("/snippet/create", app.snippetCreate)
	srv := &http.Server{
		Addr:     cfg.addr,
		ErrorLog: errorLog,
		Handler:  mux,
	}

	infoLog.Printf("starting server on %s", cfg.addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

// for a given DSN.
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
