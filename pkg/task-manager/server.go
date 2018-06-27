// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package task_manager

import (
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
	"net/http"
	"fmt"
)

func Serve(ip, port string) {
	router := httprouter.New()

	mux := http.NewServeMux()
	mux.Handle("/", router)

	handler := cors.Default().Handler(mux)
	// Todo: graceful shutdown
	err := http.ListenAndServe(fmt.Sprintf("%s:%s", ip, port), handler)
	// Todo: logging
	fmt.Println(err)
}
