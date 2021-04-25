package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

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

type Response struct {
	Env     string
	Project string
	Time    string
	Data    interface{}
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

func getCurrentEnvIndex(project_config PConfig, env_name string) int {
	env_index := 0
	for i, env_config := range project_config.EnvDetais {
		if env_config.Name == env_name {
			env_index = i
			break
		}
	}
	return env_index
}

func GetIssueDetails(w http.ResponseWriter, r *http.Request) {
	name := "get issue details"
	config, _ := getConfig(name)
	params := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")
	project_config, _ := getProjectConfig(params["proj"])
	env_index := getCurrentEnvIndex(project_config, params["env"])
	jql := "project=\"" + project_config.ProjectName + "\" AND (status=\"" + project_config.EnvDetais[env_index].Status + "\")"
	config.Url = config.Url + "?jql=" + url.QueryEscape(jql)
	resp, _ := MakeApiCall(config, nil)
	i, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err.Error())
	}
	defer resp.Body.Close()

	var responseObject IssueList
	json.Unmarshal(i, &responseObject)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	json.NewEncoder(w).Encode(Response{
		Env:     params["env"],
		Project: params["proj"],
		Time:    params["time"],
		Data:    responseObject})
}

func GetJiraProjects(w http.ResponseWriter, r *http.Request) {
	name := "get jira projects"
	config, _ := getConfig(name)
	resp, _ := MakeApiCall(config, nil)
	p, err := ioutil.ReadAll(resp.Body)
	fmt.Print(string(p))
	if err != nil {
		fmt.Print(err.Error())

	}
	defer resp.Body.Close()
	var responseObject []Project
	json.Unmarshal(p, &responseObject)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	json.NewEncoder(w).Encode(responseObject)
}

func mergeBranchesByIssueId(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var issues []Issue
	w.Header().Set("Content-Type", "application/json")
	json.NewDecoder(r.Body).Decode(&issues)
	branches := GetAllBranchesName(issues)
	fmt.Print(branches)

	apiName := "merge branch"
	config, _ := getConfig(apiName)
	pconfig, _ := getProjectConfig(params["proj"])
	env_index := getCurrentEnvIndex(pconfig, params["env"])

	for _, branch := range branches {
		fmt.Print(branch)
		postBody, _ := json.Marshal(map[string]string{
			"base": pconfig.EnvDetais[env_index].Branch,
			"head": branch,
		})
		config.Url = fmt.Sprintf(config.Url, pconfig.Owner, pconfig.Repo)
		responseBody := bytes.NewBuffer(postBody)
		resp, err := MakeApiCall(config, responseBody)
		response, _ := ioutil.ReadAll(resp.Body)
		fmt.Print(string(response))
		if err != nil {
			fmt.Print(err.Error())
		}

		if resp.StatusCode != 201 {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(err)
		}
		defer resp.Body.Close()
	}
	updateJiraTasks(issues, pconfig, params)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{
		Env:     params["env"],
		Project: params["proj"],
		Time:    params["time"],
		Data:    issues})
}

func GetAllBranchesName(issues []Issue) []string {
	name := "get branch name"
	config, _ := getConfig(name)
	params := make(map[string]string)
	fmt.Printf("length %d", len(issues))
	var channelMain = make(chan []byte, len(issues))
	var channelError = make(chan error, len(issues))

	branches := make([]string, 0)
	var wg sync.WaitGroup
	for _, issue := range issues {
		params["issueId"] = issue.ID
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
			//fmt.Printf("response %v", responseObject)
		case err := <-channelError:
			fmt.Print(err)
		default:
			done = true
		}
	}

	return branches
}

func updateJiraTasks(issues []Issue, pconfig PConfig, params map[string]string) {
	for _, issue := range issues {
		updateComment(issue.ID, params["env"])
		tran_id := getTransitions(issue.ID, pconfig, params["env"])
		updateTransition(issue.ID, tran_id)
	}
}

func updateTransition(issue_id string, tran_id string) {
	get_transitions := "transitions"
	config, _ := getConfig(get_transitions)
	config.Url = fmt.Sprintf(config.Url, issue_id)
	config.Method = "POST"
	postBody, _ := json.Marshal(map[string]map[string]string{
		"transition": {
			"id": tran_id,
		},
	})
	responseBody := bytes.NewBuffer(postBody)
	_, err := MakeApiCall(config, responseBody)
	if err != nil {
		fmt.Println("failed to update trasnsition for issue id " + issue_id)
	}
}

func getTransitions(id string, pconfig PConfig, env string) string {
	type transitionResponse struct {
		Expand      string `json:"expand"`
		Transitions []struct {
			Id   string `json:"id"`
			Name string `json:"name"`
		} `json:"transitions"`
	}
	get_transitions := "transitions"
	config, _ := getConfig(get_transitions)
	config.Url = fmt.Sprintf(config.Url, id)
	resp, _ := MakeApiCall(config, nil)
	p, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err.Error())
	}
	defer resp.Body.Close()
	var ts transitionResponse
	json.Unmarshal(p, &ts)
	var next string
	for _, e := range pconfig.EnvDetais {
		if e.Name == env {
			next = e.Next
		}
	}
	var transition string
	for _, t := range ts.Transitions {
		if t.Name == next {
			transition = t.Id
		}
	}
	return transition
}

func updateComment(id string, env string) {
	add_comment := "add jira comment"
	config, _ := getConfig(add_comment)
	config.Url = fmt.Sprintf(config.Url, id)
	postBody, _ := json.Marshal(map[string]string{
		"body": "This is updated to " + env + " by smartploy on " + time.Now().String(),
	})
	responseBody := bytes.NewBuffer(postBody)
	_, err := MakeApiCall(config, responseBody)
	if err != nil {
		fmt.Println("failed to update comment for issue id " + id)
	}
}
