package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
)

func MakeApiCallAsync(config Config, body io.Reader, params map[string]string, wg *sync.WaitGroup, channelMain chan []byte, channelError chan error) {

	defer (*wg).Done()
	client := &http.Client{}

	req, err := http.NewRequest(config.Method, config.Url, body)
	q := req.URL.Query()
	for key, value := range params {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()
	if err != nil {
		fmt.Print(err.Error())
		channelError <- err
	}

	for _, header := range config.Headers {
		req.Header.Add(header.Key, header.Value)
	}
	fmt.Print(req)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Print(err.Error())
		channelError <- err
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err.Error())
		channelError <- err
	}

	channelMain <- bodyBytes

}

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
