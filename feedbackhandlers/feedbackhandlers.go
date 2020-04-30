// The package to work with client responses

package feedbackhandlers

import (
	"log"

	// "reflect"

	Types "../types"
	"github.com/docker/docker/api/types"
	"github.com/mitchellh/mapstructure"
)

// Handles the the feedback from a client
func HandleClientFeedback(client *Types.Client, clientResponseBody []byte) error {
	response, respError := Types.UnmarshalResponseStructure(clientResponseBody)
	if respError != nil {
		log.Printf("Error in response unmarshaling")
	}
	log.Printf("The client response =  %s : %v", response.Request, response.Arguments)

	var err error
	err = nil
	switch response.Request {
	case "Status":
		err = HandleStatusFeedback(client, response.Arguments)
		break
	default:
		log.Printf("TODO: Make a handler for the FeedBackType: %s", response.Request)
		break
	}

	return err
}

// HandleStatusFeedback handles the FB from a client about a `Status` request
func HandleStatusFeedback(client *Types.Client, arguments interface{}) error {
	args := arguments.(map[string]interface{})

	// Update the image list
	client.ImageList = []types.ImageSummary{}
	if args["Image.List"] != nil {
		imageList := args["Image.List"].([]interface{})
		for _, img := range imageList {
			var image types.ImageSummary
			mapstructure.Decode(img, &image)
			client.ImageList = append(client.ImageList, image)
		}
	}

	// Update the container list
	client.ContainerList = []types.Container{}
	if args["Container.List"] != nil {
		containerList := args["Container.List"].([]interface{})
		for _, container := range containerList {
			var cont types.Container
			mapstructure.Decode(container, &cont)
			client.ContainerList = append(client.ContainerList, cont)
		}
	}

	return nil
}
