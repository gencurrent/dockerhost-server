package handlers

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	Types "../types"
)

// Разница двух строк
func difference(a, b []string) []string {

	target := map[string]bool{}
	for _, x := range b {
		target[x] = true
	}

	result := []string{}
	for _, x := range a {
		if _, ok := target[x]; !ok {
			result = append(result, x)
		}
	}

	return result
}

var HostImageList []string

func updateDockerList() {

	resp, err := http.Get("http://localhost:5000/v2/_catalog")
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var imageList map[string][]string
	err = json.Unmarshal(body, &imageList)
	if err != nil {
		log.Printf("The decoded body : %s", string(body))
		panic(err)
	}

	HostImageList = imageList["repositories"]
	log.Printf("updateDockerList.Result :: %v", HostImageList)
}

//
func DockerList() (string, error) {
	updateDockerList()
	req := Types.RequestStructure{
		"Image.List.Save",
		map[string]interface{}{
			"list": HostImageList,
		},
	}
	return req.Marshal()
}

// RequestToPullImage a client to push images
func RequestToPullImage(imageList []string) (string, error) {
	updateDockerList()
	diff := difference(HostImageList, imageList)
	req := Types.RequestStructure{
		`Image.Pull`,
		map[string]interface{}{
			"List": diff,
		},
	}
	return req.Marshal()
}

func Status() (string, error) {
	request := Types.RequestStructure{
		Request: "Status",
		Arguments: map[string]interface{}{
			"put": []string{"Image.List", "Container.List"},
		},
	}
	res, err := request.Marshal()
	return res, err
}

func Idle() (string, error) {
	return `{"request": null, "arguments":{}}`, nil
}