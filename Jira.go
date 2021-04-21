package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
)

func init() {
	r := []route{
		{
			Method:  "GET",
			Path:    "/get_jira_projects",
			Handler: GetJiraProjects,
		},
		{
			Method:  "POST",
			Path:    "/get_issues",
			Handler: GetIssueDetails,
		},
		{
			Method:  "POST",
			Path:    "/merge_by_issesId",
			Handler: mergeBrancesByIssueId,
		},
	}
	registerRoute(r)
}

type BranchesJson struct {
	Details []struct {
		Branches []struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"branches"`
	} `json:"detail"`
}
type Project struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

type Issue struct {
	ID     string       `json:"id"`
	Key    string       `json:"key"`
	Fields *IssueFields `json:"fields"`
}

type IssueList struct {
	Total  int     `json:"total"`
	Issues []Issue `json:"issues"`
}

type IssueFields struct {
	Summary string `json:"summary"`
	Parent  *Issue `json:"parent"`
}

func GetIssueDetails(w http.ResponseWriter, r *http.Request) {
	name := "get issue details"
	config, _ := getConfig(name)
	type req struct {
		Key string `json:"key"`
		Id  int    `json:"id"`
		Env string `json:"env"`
	}
	fmt.Println(config)
	var params req
	w.Header().Set("Content-Type", "application/json")
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		fmt.Println(err)
		return
	}
	esm := map[string]string{
		"STG":     "CODE REVIEW",
		"UAT":     "TESTED IN STG",
		"PREPROD": "TESTED IN UAT",
		"PROD":    "TESTED IN PREPROD",
	}
	jql := "project=\"" + params.Key + "\" AND (status=\"" + esm[params.Env] + "\")"
	config.Url = config.Url + "?jql=" + url.QueryEscape(jql)
	fmt.Println(config)
	i, _ := MakeApiCall(config, nil)
	fmt.Println(string(i))
	var responseObject IssueList
	json.Unmarshal(i, &responseObject)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseObject)
}

func GetJiraProjects(w http.ResponseWriter, r *http.Request) {
	name := "get jira projects"
	config, _ := getConfig(name)
	p, _ := MakeApiCall(config, nil)
	var responseObject []Project
	json.Unmarshal(p, &responseObject)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseObject)
}

func mergeBrancesByIssueId(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	issues := make([]string, 0)
	values := r.Form
	for value := range values {
		issues = append(issues, value)
	}

	// branches :=
	GetAllBranchesName(issues)
	//name :=
}

func GetAllBranchesName(issues []string) []string {
	name := "get branch name"
	config, _ := getConfig(name)
	params := make(map[string]string)
	var channelMain = make(chan []byte, len(issues))
	var channelError = make(chan error, len(issues))

	branches := make([]string, 0)
	var wg sync.WaitGroup
	for _, issue := range issues {
		params["issueId"] = issue
		params["applicationType"] = "GitHub"
		params["dataType"] = "branch"
		wg.Add(1)
		go MakeApiCallAsync(config, nil, params, &wg, channelMain, channelError)

	}
	wg.Wait()

	for done := false; !done; {
		select {
		case response := <-channelMain:
			var responseObject BranchesJson
			json.Unmarshal(response, &responseObject)
			branches = append(branches, responseObject.Details[0].Branches[0].Name)
		case err := <-channelError:
			fmt.Print(err)
		default:
			done = true
		}
	}

	return branches
}
