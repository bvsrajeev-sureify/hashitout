package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type Config struct {
	Url     string   `json:"url"`
	Method  string   `json:"method"`
	Name    string   `json:"name"`
	Headers []Header `json:"headers"`
}

type PConfig struct {
	ProjectName string `json:"project_name"`
	Repo        string `json:"repository"`
	Owner       string `json:"owner"`
	EnvDetais   []struct {
		Name   string `json:"env_name"`
		Status string `json:"issue_status"`
		Branch string `json:"env_branch"`
		Next   string `json:"next_status"`
	} `json:"env_config"`
}

type Header struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

var ApiConfig []Config

var ProjectConfig []PConfig

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

func updateConfig(config string) []byte {
	for key, value := range Env {
		config = strings.ReplaceAll(config, "$"+key, value)
	}
	return []byte(config)
}

func (cf *ConfigFile) Read() ([]byte, error) {
	config, err := ioutil.ReadFile(cf.name)
	config = updateConfig(string(config))
	return config, err

}

var routes = make([]route, 0)

func registerRoute(r []route) {
	routes = append(routes, r...)
}

var Env map[string]string

func loadEnvironment() {
	env, _ := ioutil.ReadFile("env.json")
	json.Unmarshal(env, &Env)
}

func main() {
	// Route handles & endpoints
	loadEnvironment()
	r := mux.NewRouter()
	data := ConfigFile{"config.json"}
	projects := ConfigFile{"project.json"}

	config, err := data.Read()
	projectConfig, _ := projects.Read()

	if err == nil {
		json.Unmarshal(config, &ApiConfig)
	}
	json.Unmarshal(projectConfig, &ProjectConfig)
	fmt.Println((ApiConfig))

	// handlers.AllowedHeaders([]string{"X-Requested-With"})
	// handlers.AllowedOrigins([]string{"*"})
	// handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	for _, rt := range routes {
		r.HandleFunc(rt.Path, rt.Handler).Methods(rt.Method)
	}
	log.Fatal(http.ListenAndServe(":8000", handlers.CORS(handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}), handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS"}), handlers.AllowedOrigins([]string{"*"}))(r)))
}
