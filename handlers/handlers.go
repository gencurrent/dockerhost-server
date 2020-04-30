/***
The handlers for the every request type from server to a client
*/

package handlers

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	Types "../types"
	Utils "../utils"
)

var HostImageList []string

func UpdateLocalImageList() {

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

func RequestToRunImage(imageName string) (string, error) {
	req := Types.RequestStructure{
		`Image.Run`,
		map[string]interface{}{
			"Image.Name": imageName,
		},
	}
	return req.Marshal()
}

func RequestToStopContainer(containerID string) (string, error) {
	req := Types.RequestStructure{
		`Container.Stop`,
		map[string]interface{}{
			"Container.ID": containerID,
		},
	}
	return req.Marshal()
}

// RequestToPullImage a client to pull images
func RequestToPullImage(imageList []string) (string, error) {
	UpdateLocalImageList()
	diff := Utils.Difference(HostImageList, imageList)
	req := Types.RequestStructure{
		`Image.Pull`,
		map[string]interface{}{
			"List": diff,
		},
	}
	return req.Marshal()
}

// Status : check the Client status
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
