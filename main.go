package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Config struct {
	Url     string   `json:"url"`
	Method  string   `json:"method"`
	Name    string   `json:"name"`
	Headers []Header `json:"headers"`
}

type Header struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

var ApiConfig []Config

type route struct {
	Method  string
	Path    string
	Handler func(w http.ResponseWriter, r *http.Request)
}

type Reader interface {
	Read()
}

type ConfigFile struct {
	name string
}

func (cf *ConfigFile) Read() ([]byte, error) {
	config, err := ioutil.ReadFile(cf.name)
	return config, err

}

var routes = make([]route, 0)

func registerRoute(r []route) {
	routes = append(routes, r...)
}

func getConfig(name string) (Config, error) {
	var config Config
	for _, c := range ApiConfig {
		if c.Name == name {
			config = c
		}
	}
	return config, nil
}

func main() {
	// Route handles & endpoints
	r := mux.NewRouter()
	data := ConfigFile{"config.json"}

	config, err := data.Read()
	if err == nil {
		json.Unmarshal(config, &ApiConfig)
	}

	for _, rt := range routes {
		r.HandleFunc(rt.Path, rt.Handler).Methods(rt.Method)
	}
	log.Fatal(http.ListenAndServe(":8000", r))
}
