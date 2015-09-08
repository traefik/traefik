package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"github.com/gorilla/handlers"
	"github.com/unrolled/render"
	"fmt"
	"html/template"
)

var renderer = render.New()

type WebProvider struct {
	Address string
}

type Page struct {
	Configuration Configuration
}

func (provider *WebProvider) Provide(configurationChan chan<- *Configuration){
	systemRouter := mux.NewRouter()
	systemRouter.Methods("GET").PathPrefix("/html/").Handler(handlers.CombinedLoggingHandler(os.Stdout, http.HandlerFunc(GetHtmlConfigHandler)))
	systemRouter.Methods("GET").PathPrefix("/json/").Handler(handlers.CombinedLoggingHandler(os.Stdout, http.HandlerFunc(GetConfigHandler)))
	systemRouter.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	go http.ListenAndServe(provider.Address, systemRouter)
}

func GetConfigHandler(rw http.ResponseWriter, r *http.Request) {
	renderer.JSON(rw, http.StatusOK, currentConfiguration)
}

func GetHtmlConfigHandler(response http.ResponseWriter, request *http.Request) {
	templates := template.Must(template.ParseFiles("configuration.html"))
	response.Header().Set("Content-type", "text/html")
	err := request.ParseForm()
	if err != nil {
		http.Error(response, fmt.Sprintf("error parsing url %v", err), 500)
	}
	templates.ExecuteTemplate(response, "configuration.html", Page{Configuration:*currentConfiguration})
}
