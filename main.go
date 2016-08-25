package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/handlers"
	"github.com/jessevdk/go-flags"
	"github.com/julienschmidt/httprouter"
	"github.com/oschwald/geoip2-golang"
)

var opts struct {
	// Example of a required flag
	DBPath string `long:"dbPath" description:"path to the maxmind DB" required:"true"`

	// Example of a pointer
	Port int `long:"port" default:"80" description:"Port for HTTP API"`

	Logfile    string `long:"logFile" description:"path to write logs to"`
	LogBacklog int    `long:"logBacklog" default:"1000" description:"Buffer size for async log"`
}

func main() {
	parser := flags.NewParser(&opts, flags.Default)
	if _, err := parser.Parse(); err != nil {
		logrus.Fatalf("Error parsing flags: %v", err)
	}

	// Open up the maxmind DB
	db, err := geoip2.Open(opts.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create the HTTP endpoint
	router := httprouter.New()
	httpApi := NewHTTPApi(db)
	httpApi.Start(router)

	var loggingHandler http.Handler
	if opts.Logfile != "" {
		fh, err := os.OpenFile(opts.Logfile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			logrus.Fatalf("unable to open log file: %v", err)
		}
		a := NewAsyncLogWriter(fh, opts.LogBacklog)
		loggingHandler = handlers.LoggingHandler(a, router)
	} else {
		loggingHandler = handlers.LoggingHandler(os.Stdout, router)
	}

	if err := http.ListenAndServe(fmt.Sprintf(":%d", opts.Port), loggingHandler); err != nil {
		logrus.Fatalf("Error serving httpapi: %v", err)
	}
}
