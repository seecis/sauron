// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package http

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"log"
	"fmt"
	"mime"
	"encoding/json"
	"gopkg.in/yaml.v2"
	"github.com/seecis/sauron/internal/dataaccess"
)

func StartServer(ip, port string) {

	fses := &dataaccess.FileSystemExtractorService{}
	eh := &ExtractorHandler{service: fses}
	router := httprouter.New()
	router.GET("/", index)
	router.GET("/extractor", eh.GetAll)
	router.PUT("/extractor", eh.NewExtractor)

	router.GET("/extractor/:id", index)
	router.DELETE("/extractor/:id", index)
	router.POST("/extractor/:id", index)

	router.POST("/extract/:id", index)

	router.GET("/report/:id", index)

	address := fmt.Sprintf("%s:%s", ip, port)
	log.Printf("Sauron is listening you at %s", address)
	log.Fatal(http.ListenAndServe(address, router))
}

func index(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	fmt.Fprintln(w, "Sauron is watching you")
	fmt.Fprintln(w, "You requested "+r.URL.Path)
	fmt.Fprintln(w, params)
}

type responseType int

const (
	mime_yaml responseType = iota
	mime_json
)

func getResponseType(r *http.Request) responseType {
	// todo: test this EXHAUSTIVELY

	s := r.Header.Get("Accept")
	log.Println("Accept header is ", s)
	mediaType, params, err := mime.ParseMediaType(s)
	if err != nil {
		return mime_yaml
	}

	log.Println(params, mediaType)

	switch mediaType {
	case "application/json":
		return mime_json
	case "application/yaml":
		return mime_yaml
	}
	return mime_yaml
}

type ExtractorHandler struct {
	service dataaccess.ExtractorService
}

func (eh *ExtractorHandler) GetAll(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	res := getResponseType(r)
	allExtractors, err := eh.service.GetAll()
	if err != nil {
		log.Panic(err)
	}

	switch res {
	case mime_json:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(allExtractors)
	case mime_yaml:
		yaml.NewEncoder(w).Encode(allExtractors)
	}

	return
}

func (eh *ExtractorHandler) NewExtractor(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

}

func (eh *ExtractorHandler) GetExtractor(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

}

func (eh *ExtractorHandler) DeleteExtractor(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

}

func (eh *ExtractorHandler) UpdateExtractor(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

}

/*
/extractor
	create
		Accepts: json, yaml
	read
		Produces: json, yaml
/extractor/{id}
	read
		Produces: json, yaml
	delete
		Just delete
	update
		Accepts: json, yaml
		Does not return the extractor
/extract?extractorId={"extractorId"}
	POST
	Accepts: json, yaml
	Example:
	{
		"url":["example.com"]
	}
	Returns: [json, yaml], Created report Id, status
/report/{reportId}
	create ??
	read
	update ??
	delete
*/
