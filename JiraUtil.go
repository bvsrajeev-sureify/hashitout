package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"time"
)

type Jira struct {
	project PConfig
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

func (j *Jira) GetProjects() []Project {
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
	return responseObject
}

func (j *Jira) GetIssues(project string, env string) IssueList {
	name := "get issue details"
	config, _ := getConfig(name)
	j.getProjectConfig(project)
	env_index := j.getCurrentEnvIndex(env)
	jql := "project=\"" + j.project.ProjectName + "\" AND (status=\"" + j.project.EnvDetais[env_index].Status + "\")"
	config.Url = config.Url + "?jql=" + url.QueryEscape(jql)
	resp, _ := MakeApiCall(config, nil)
	i, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err.Error())
	}
	defer resp.Body.Close()

	var responseObject IssueList
	json.Unmarshal(i, &responseObject)
	return responseObject
}

func (j *Jira) getCurrentEnvIndex(env_name string) int {
	env_index := 0
	for i, env_config := range j.project.EnvDetais {
		if env_config.Name == env_name {
			env_index = i
			break
		}
	}
	return env_index
}

func (j *Jira) GetAllBranchesName(issues []Issue) []string {
	name := "get branch name"
	// config, _ := getConfig(name)
	// params := make(map[string]string)
	// fmt.Printf("length %d", len(issues))
	// var channelMain = make(chan []byte, len(issues))
	// var channelError = make(chan error, len(issues))

	type BranchesJson struct {
		Details []struct {
			Branches []struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"branches"`
		} `json:"detail"`
	}

	branches := make([]string, 0)
	// var wg sync.WaitGroup
	for _, issue := range issues {
		// params["issueId"] = issue.ID
		config, _ := getConfig(name)
		config.Url = config.Url + "?issueId=" + issue.ID
		// wg.Add(1)
		resp, _ := MakeApiCall(config, nil)
		response, _ := ioutil.ReadAll(resp.Body)
		var responseObject BranchesJson
		json.Unmarshal(response, &responseObject)
		branches = append(branches, responseObject.Details[0].Branches[0].Name)

	}
	// wg.Wait()

	// for done := false; !done; {
	// 	select {
	// 	case response := <-channelMain:
	// 		// var responseObject BranchesJson
	// 		// json.Unmarshal(response, &responseObject)
	// 		// branches = append(branches, responseObject.Details[0].Branches[0].Name)
	// 		//fmt.Printf("response %v", responseObject)
	// 	case err := <-channelError:
	// 		fmt.Print(err)
	// 	default:
	// 		done = true
	// 	}
	// }

	return branches
}

func (j *Jira) getProjectConfig(name string) {
	for _, c := range ProjectConfig {
		if c.ProjectName == name {
			j.project = c
		}
	}
}

func (j *Jira) UpdateJiraTasks(issues []Issue, env string) {
	for _, issue := range issues {
		j.updateComment(issue.ID, env)
		tran_id := j.getTransitions(issue.ID, env)
		j.updateTransition(issue.ID, tran_id)
	}
}

func (j *Jira) updateTransition(issue_id string, tran_id string) {
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

func (j *Jira) getTransitions(id string, env string) string {
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
	for _, e := range j.project.EnvDetais {
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

func (j *Jira) updateComment(id string, env string) {
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
