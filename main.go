package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"
	"github.com/julienschmidt/httprouter"
	"github.com/oschwald/geoip2-golang"
)

var opts struct {
	// Example of a required flag
	DBPath string `long:"dbPath" description:"path to the maxmind DB" required:"true"`

	// Example of a pointer
	Port int `long:"port" default:"80" description:"Port for HTTP API"`
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

	if err := http.ListenAndServe(fmt.Sprintf(":%d", opts.Port), router); err != nil {
		logrus.Fatalf("Error serving httpapi: %v", err)
	}
}
