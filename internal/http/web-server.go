// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package http

import (
	"net/http"
	"html/template"
	"github.com/julienschmidt/httprouter"
	"fmt"
	"log"
)

func ServeWeb(ip, port string) error {
	router := httprouter.New()
	router.Handle("GET", "/", mainHandler)
	router.Handle("GET", "/extractor", extractorsHandler)
	addr := fmt.Sprintf("%s:%s", ip, port)
	log.Printf("Listening on %s", addr)
	return http.ListenAndServe(addr, router)
}

func extractorsHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	t, _ := template.ParseFiles("templates/extractor.html")
	t.Execute(w, "")
}

func mainHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	t, _ := template.ParseFiles("templates/view.html")
	t.Execute(w, "Hello world")
}
