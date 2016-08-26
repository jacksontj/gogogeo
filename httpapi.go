package main

import (
	"encoding/json"
	"net"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/oschwald/geoip2-golang"
	"github.com/oschwald/maxminddb-golang"
)

type HTTPApi struct {
	dbMap map[string]*geoip2.Reader
}

type ipLookup interface {
	Test(net.IP) (interface{}, error)
}

func NewHTTPApi(dbs []*geoip2.Reader) *HTTPApi {
	api := &HTTPApi{
		dbMap: make(map[string]*geoip2.Reader),
	}

	// TODO: with reflect? Its lame to have to hard code all this
	for _, db := range dbs {
		var err error
		_, err = db.AnonymousIP(nil)
		if _, ok := err.(geoip2.InvalidMethodError); !ok {
			api.setDB("anonymousip", db)
		}

		_, err = db.City(nil)
		if _, ok := err.(geoip2.InvalidMethodError); !ok {
			api.setDB("city", db)
		}

		_, err = db.ConnectionType(nil)
		if _, ok := err.(geoip2.InvalidMethodError); !ok {
			api.setDB("connectiontype", db)
		}

		_, err = db.Country(nil)
		if _, ok := err.(geoip2.InvalidMethodError); !ok {
			api.setDB("country", db)
		}

		_, err = db.Domain(nil)
		if _, ok := err.(geoip2.InvalidMethodError); !ok {
			api.setDB("domain", db)
		}

		_, err = db.ISP(nil)
		if _, ok := err.(geoip2.InvalidMethodError); !ok {
			api.setDB("isp", db)
		}
	}

	return api
}

func (h *HTTPApi) setDB(k string, db *geoip2.Reader) {
	if _, ok := h.dbMap[k]; !ok {
		h.dbMap[k] = db
	}
}

func (h *HTTPApi) Start(router *httprouter.Router) {
	router.GET("/v1/metadata", h.metadataHandler)
	router.GET("/v1/geo/:kind/:ip", h.ipHandler)
}

func (h *HTTPApi) ipHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ip := net.ParseIP(ps.ByName("ip"))

	db, ok := h.dbMap[ps.ByName("kind")]
	if !ok {
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	var item interface{}
	var err error

	switch ps.ByName("kind") {
	case "anonymousip":
		item, err = db.AnonymousIP(ip)
	case "city":
		item, err = db.City(ip)
	case "connectiontype":
		item, err = db.ConnectionType(ip)
	case "country":
		item, err = db.Country(ip)
	case "domain":
		item, err = db.Domain(ip)
	case "isp":
		item, err = db.ISP(ip)
	default:
		http.NotFound(w, r)
		return
	}

	// if the IP was bad-- 400!
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ret, err := json.Marshal(item)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(ret)
	}
}

func (h *HTTPApi) metadataHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	m := make(map[string]maxminddb.Metadata)
	for k, db := range h.dbMap {
		m[k] = db.Metadata()
	}
	ret, err := json.Marshal(m)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(ret)
	}
}
