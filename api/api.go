package api

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func Start() {

	router := mux.NewRouter()
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./api/static/")))
	log.Fatal(http.ListenAndServe(":8100", router))
}
