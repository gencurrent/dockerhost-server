package handlers

import (
	"io/ioutil"
	"net/http"
)

func DockerList() (string, error) {
	resp, err := http.Get("http://localhost:5000/v2/_catalog")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return string(body), nil
}

func RequestToPullImage() (string, error) {
	return `{"request":"image.pull", "arguments": {"name": "192.168.1.63:5000/postgres"}}`, nil
}

func Idle() (string, error) {
	return `{"request": null, "arguments":{}}`, nil
}
