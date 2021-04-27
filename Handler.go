package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func init() {
	r := []route{
		{
			Method:  "GET",
			Path:    "/get_jira_projects",
			Handler: GetJiraProjects,
		},
		{
			Method:  "GET",
			Path:    "/get_issues/{proj}/{env}/{time}",
			Handler: GetIssueDetails,
		},
		{
			Method:  "POST",
			Path:    "/merge_by_issesId/{proj}/{env}/{time}",
			Handler: mergeBranchesByIssueId,
		},
	}
	registerRoute(r)
}

type Response struct {
	Env     string
	Project string
	Time    string
	Data    interface{}
}

var jira Jira
var github Github

func getConfig(name string) (Config, error) {
	var config Config
	for _, c := range ApiConfig {
		if c.Name == name {
			config = c
		}
	}
	return config, nil
}

func GetIssueDetails(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	responseObject := jira.GetIssues(params["proj"], params["env"])
	json.NewEncoder(w).Encode(Response{
		Env:     params["env"],
		Project: params["proj"],
		Time:    params["time"],
		Data:    responseObject})
}

func GetJiraProjects(w http.ResponseWriter, r *http.Request) {
	responseObject := jira.GetProjects()
	json.NewEncoder(w).Encode(responseObject)
}

func mergeBranchesByIssueId(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var issues []Issue
	json.NewDecoder(r.Body).Decode(&issues)
	branches := jira.GetAllBranchesName(issues)
	fmt.Println(branches)

	jira.getProjectConfig(params["proj"])
	env_index := jira.getCurrentEnvIndex(params["env"])

	success := github.MergeBranches(branches, jira.project, env_index)
	if !success {
		json.NewEncoder(w).Encode("error")
	}
	jira.UpdateJiraTasks(issues, params["env"])
	json.NewEncoder(w).Encode(Response{
		Env:     params["env"],
		Project: params["proj"],
		Time:    params["time"],
		Data:    issues})
}
