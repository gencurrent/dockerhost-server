package main

import (
	"testing"

	FeedbackHandlers "./feedbackhandlers"
	Types "./types"
)

// TestClient — test a client functions
func TestClient(t *testing.T) {
	ip := "127.0.0.1"
	client := Types.NewClient(
		ip,
		"646",
	)
	if client.IP != ip {
		t.Errorf("The Client.IP is not same: %s != %s", client.IP, ip)
	}
}

func TestFeedbackHandling(t *testing.T) {
	ip := "127.0.0.1"
	client := Types.NewClient(
		ip,
		"646",
	)
	clientMessageBody := `{"request": "Status", "arguments": {"List": ["postgres", "alpine"]} }`
	result := FeedbackHandlers.HandleClientFeedback(client, []byte(clientMessageBody))
	if result != nil {
		t.Errorf("Awful!")
	}
	t.Logf("Client response is handled correctly")
	imageList := []string{"postgres", "alpine"}
	if len(imageList) != len(client.ImageList) {
		t.Errorf("The images could not be set to the client")
		return
	}
	for i, img := range client.ImageList {
		if img != imageList[i] {
			t.Errorf("TestFeedbackHandling:: Wrong image set: %s, %s", img, imageList[i])
		}
	}
	t.Logf("Client image list is matched")
}
