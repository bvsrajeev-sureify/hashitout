package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Github struct {
}

func (g *Github) MergeBranches(branches []string, project PConfig, env_index int) bool {
	apiName := "merge branch"
	for _, branch := range branches {
		config, _ := getConfig(apiName)
		fmt.Println(branch)
		postBody, _ := json.Marshal(map[string]string{
			"base": project.EnvDetais[env_index].Branch,
			"head": branch,
		})
		config.Url = fmt.Sprintf(config.Url, project.Owner, project.Repo)
		responseBody := bytes.NewBuffer(postBody)
		resp, err := MakeApiCall(config, responseBody)
		if err != nil {
			fmt.Print(err.Error())
		}
		defer resp.Body.Close()
		response, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(response))

		if resp.StatusCode != 201 {
			return false
		}
	}
	return true
}
