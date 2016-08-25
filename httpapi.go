package main

import (
	"encoding/json"
	"net"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/oschwald/geoip2-golang"
)

type HTTPApi struct {
	db *geoip2.Reader
}

func NewHTTPApi(db *geoip2.Reader) *HTTPApi {
	api := &HTTPApi{
		db: db,
	}

	return api
}

func (h *HTTPApi) Start(router *httprouter.Router) {
	router.GET("/v1/metadata", h.metadataHandler)
	router.GET("/v1/geo/:kind/:ip", h.ipHandler)
}

func (h *HTTPApi) ipHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ip := net.ParseIP(ps.ByName("ip"))

	var item interface{}
	var err error

	switch ps.ByName("kind") {
	case "anonymousip":
		item, err = h.db.AnonymousIP(ip)
	case "city":
		item, err = h.db.City(ip)
	case "connectiontype":
		item, err = h.db.ConnectionType(ip)
	case "country":
		item, err = h.db.Country(ip)
	case "domain":
		item, err = h.db.Domain(ip)
	case "isp":
		item, err = h.db.ISP(ip)
	case "metadata":
		item = h.db.Metadata()
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
	ret, err := json.Marshal(h.db.Metadata())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(ret)
	}
}
