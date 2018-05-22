// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

///extractor
//	create
//		Accepts: json, yaml
//	read
//		Produces: json, yaml
///extractor/{id}
//	read
//		Produces: json, yaml
//	delete
//		Just delete
//	update
//		Accepts: json, yaml
//		Does not return the extractor
///extract?extractorId={"extractorId"}
//	POST
//	Accepts: json, yaml
//	Example:
//	{
//		"url":["example.com"]
//	}
//	Returns: [json, yaml], Created report Id, status
///report/{reportId}
//	create ??
//	read
//	update ??
//	delete

package http

import (
	"encoding/json"
	"fmt"
	"github.com/seecis/sauron/internal/dataaccess"
	"github.com/seecis/sauron/pkg/extractor"
	"github.com/seecis/sauron/pkg/scheduler"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"github.com/gorilla/handlers"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"log"
	"mime"
	"net/http"
	"strconv"
)

func ServeApi(ip, port string) {

	// todo add this to config.
	//fses := dataaccess.NewFileSystemExtractorService("extractors/")
	mses := dataaccess.NewMsSqlExtractorService(true, false)
	eh := &ExtractorHandler{service: mses}

	router := httprouter.New()
	router.GET("/", index)
	router.GET("/extractor", eh.GetAll)
	router.PUT("/extractor", eh.NewExtractor)

	router.GET("/extractor/:id", eh.GetExtractor)
	router.DELETE("/extractor/:id", eh.DeleteExtractor)
	router.POST("/extractor/:id", eh.UpdateExtractor)
	router.POST("/extract/:id", eh.Extract)
	router.GET("/report/:id", eh.GetReport)

	address := fmt.Sprintf("%s:%s", ip, port)
	log.Printf("Sauron is listening you at %s", address)

	accessLogFile := createAccessLog()
	defer accessLogFile.Close()
	multiout := io.MultiWriter(accessLogFile, os.Stdout)

	log.Fatal(http.ListenAndServe(address, handlers.LoggingHandler(multiout, router)))
}

func checkFileExists(path string) bool {
	//os.Stat check for file info. If Stat couldn't find a file returns an error. Since golang doesn't provide an error
	//handling mechanism os library provides us with a nice IsNotExists function. IsNotExists(error) checks if returned
	//error occurred because of an file not found error. Since we don't care about the file info we just ignore Stat's
	//first return value.
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}

func createAccessLog() *os.File {
	// Todo: add to config
	basePath := "access_log"
	extension := ".txt"

	created := false

	path := basePath + extension
	for i := 1; created == false; i++ {
		exists := checkFileExists(path)
		if exists {
			path = basePath + strconv.Itoa(i) + extension
			created = false
			continue
		}

		created = true
	}

	f, err := os.Create(path)
	if err != nil {
		// What to do here?
		log.Panic(err)
	}

	return f
}

func index(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	fmt.Fprintln(w, "Sauron is watching you")
	fmt.Fprintln(w, "You requested "+r.URL.Path)
	fmt.Fprintln(w, params)
}

type mimeType int

const (
	mime_yaml mimeType = iota
	mime_json
)

func getMimeType(typeHeader string) mimeType {
	mediaType, _, err := mime.ParseMediaType(typeHeader)
	if err != nil {
		return mime_yaml
	}

	switch mediaType {
	case "application/json":
		return mime_json
	case "application/yaml":
		return mime_yaml
	}

	return mime_yaml
}

func getResponseType(r *http.Request) mimeType {
	// todo: test this EXHAUSTIVELY
	s := r.Header.Get("Accept")
	return getMimeType(s)
}

func getContentType(r *http.Request) mimeType {
	s := r.Header.Get("Content-Type")
	return getMimeType(s)
}

type ExtractorHandler struct {
	service       dataaccess.ExtractorService
	reportService dataaccess.ReportService
	scheduler     scheduler.ExtractionScheduler
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
		w.Header().Set("Content-Type", "application/yaml")
		yaml.NewEncoder(w).Encode(allExtractors)
	}

	return
}

type Unmarshaller interface {
	Decode(interface{}) error
}

func serialize(w http.ResponseWriter, thing interface{}, mimeType mimeType) error {
	switch mimeType {
	case mime_json:
		return json.NewEncoder(w).Encode(thing)
	case mime_yaml:
		return yaml.NewEncoder(w).Encode(thing)
	}

	return errors.New("Unknown mime type")
}

func deserialize(r io.Reader, htmlExtractor *extractor.HtmlExtractor, mimeType mimeType) error {
	var unMarshaller Unmarshaller
	switch mimeType {
	case mime_json:
		unMarshaller = json.NewDecoder(r)
	case mime_yaml:
		unMarshaller = yaml.NewDecoder(r)
	}

	decode := unMarshaller.Decode(htmlExtractor)
	return decode
}

func (eh *ExtractorHandler) NewExtractor(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	typ := getContentType(r)
	// Todo: we may need more magic here
	var ex extractor.HtmlExtractor

	err := deserialize(r.Body, &ex, typ)
	defer r.Body.Close()
	if err != nil {
		// Todo: make this warning
		log.Println("Error while unmarshalling :" + err.Error())
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	id, se := eh.service.Save(&ex)
	if se != nil {
		log.Printf(se.Error())
		http.Error(w, "An error occurred", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", "/extractor/"+id)
	w.WriteHeader(http.StatusCreated)
}

func (eh *ExtractorHandler) GetExtractor(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	name := params.ByName("id")
	ex, srvErr := eh.service.Get(name)
	if srvErr != nil {
		if dataaccess.IsNotFound(srvErr) {
			http.NotFound(w, r)
			return
		}

		log.Println("Cannot get extractor with id: " + name)
		http.Error(w, "An error occurred", http.StatusInternalServerError)
		return
	}

	defer r.Body.Close()
	serialize(w, ex, getResponseType(r))
}

func (eh *ExtractorHandler) DeleteExtractor(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	extractorId := params.ByName("id")
	if extractorId == "" {
		http.Error(w, "Bad request: id is needed on request path", http.StatusBadRequest)
		return
	}

	se := eh.service.Delete(extractorId)
	if se != nil {
		if dataaccess.IsNotFound(se) {
			http.NotFound(w, r)
		}
	}
}

func (eh *ExtractorHandler) UpdateExtractor(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	name := params.ByName("id")
	if name == "" {
		http.Error(w, "Bad request: id is needed on request path", http.StatusBadRequest)
		return
	}

	ex, err := eh.service.Get(name)
	if err != nil {
		if dataaccess.IsNotFound(err) {
			http.NotFound(w, r)
			return
		}

		log.Println(err.Error())
		if dataaccess.IsBadRequest(err) {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		http.Error(w, "An error occurred", http.StatusInternalServerError)
		return
	}

	defer r.Body.Close()
	var newEx extractor.HtmlExtractor
	err = deserialize(r.Body, &newEx, getContentType(r))
	if err != nil {
		log.Println(err)
		http.Error(w, "Error while deserializing request", http.StatusBadRequest)
		return
	}

	err = eh.service.Delete(ex.GetName())
	if err != nil {
		log.Panic(err)
		return
	}

	newEx.Name = ex.GetName()
	_, err = eh.service.Save(&newEx)
	if err != nil {
		log.Panic(err)
	}
}

func (eh *ExtractorHandler) Extract(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	mm := getResponseType(r)

	extractorId := params.ByName("id")
	if extractorId == "" {
		http.Error(w, "Needs to provide an extractor id", http.StatusBadRequest)
		return
	}

	ex, err := eh.service.Get(extractorId)
	if dataaccess.IsNotFound(err) {
		http.NotFound(w, r)
		return
	}

	var um Unmarshaller
	defer r.Body.Close()
	switch mm {
	case mime_yaml:
		um = yaml.NewDecoder(r.Body)
	case mime_json:
		um = json.NewDecoder(r.Body)
	default:
		http.Error(w, "Only yaml or json allowed", http.StatusUnsupportedMediaType)
		return
	}

	var e scheduler.ExtractionRequest
	err = um.Decode(&e)
	if err != nil {
		http.Error(w, "Malformed payload", http.StatusBadRequest)
		return
	}

	reportId, err := eh.scheduler.Schedule(ex, e)
	if err != nil {
		http.Error(w, "Error while scheduling extractor for request", http.StatusInternalServerError)
		log.Println("Error while scheduling", err, e)
		return
	}

	w.Header().Set("Location", "/report/"+reportId)
	w.WriteHeader(http.StatusAccepted)
	return

}

type HtmlExtractorScheduler struct {
}

func (hes *HtmlExtractorScheduler) Schedule(extractor extractor.Extractor, payload interface{}) (string, error) {
	switch payload.(type) {
	}

	return "", nil
}

func (eh *ExtractorHandler) GetReport(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

}
