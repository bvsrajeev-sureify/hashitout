package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

func MakeApiCall(config Config, body io.Reader) ([]byte, error) {
	client := &http.Client{}

	req, err := http.NewRequest(config.Method, config.Url, body)

	if err != nil {
		fmt.Print(err.Error())
		return nil, err
	}

	for _, header := range config.Headers {
		req.Header.Add(header.Key, header.Value)
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Print(err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err.Error())
		return nil, err
	}
	return bodyBytes, nil
}
