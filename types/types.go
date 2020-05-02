package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/docker/docker/api/types"

	uuid "github.com/satori/go.uuid"
)

// Запрос от сервера
type RequestStructure struct {
	Name      string                 `json:"request"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ResponseStructure Структура ответа от клиента
type ResponseStructure struct {
	Request   string                 `json:"request"`
	Arguments map[string]interface{} `json:"arguments"`
}

const (
	ClientActive  uint8 = iota
	ClientPassive uint8 = iota
)

// Client structure defines the websocket client fields
type Client struct {
	UUID          uuid.UUID            `json:"uuid"`
	IP            string               `json:"ip"`
	Port          uint16               `json:"port"`
	ImageList     []types.ImageSummary `json:"imageList"`
	ContainerList []types.Container    `json:"containerList"`
	Status        uint8                `json:"status"`
	RequestQueue  []RequestStructure   `json:"requestQueue"`
}

// ClientList - list of clients with methods
type ClientMap struct {
	ClientList []Client
}

type Config struct {
	RegistryURL string `json:"registryURL"`
}

// marshal Encode to JSON String
func (r *RequestStructure) Marshal() (string, error) {
	result, err := json.Marshal(r)
	return string(result), err
}

// marshal Decode string to the structure type
func UnmarshalResponseStructure(responseBody []byte) (*ResponseStructure, error) {
	var responseStruct ResponseStructure
	err := json.Unmarshal(responseBody, &responseStruct)
	if err != nil {
		log.Fatal("THe error while Unmarshalling the string from client")
		log.Fatalf("%s\n", err)
		return nil, err
	}
	// return responseStruct, err
	return &responseStruct, nil
}

func NewClient(ip string, port string) *Client {
	clientPortParsed, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		panic(err)
	}
	clientPort := uint16(clientPortParsed)
	client := Client{
		IP:   ip,
		Port: clientPort,
		UUID: uuid.Must(uuid.NewV4()),
	}
	return &client
}

func (client *Client) PushRequestStructure(rs *RequestStructure) (*RequestStructure, error) {
	client.RequestQueue = append(client.RequestQueue, *rs)
	rs = &client.RequestQueue[len(client.RequestQueue)-1]
	return rs, nil
}

func (client *Client) PopRequestStructure() *RequestStructure {
	if len(client.RequestQueue) == 0 {
		return nil
	}
	rs := client.RequestQueue[0]
	client.RequestQueue = client.RequestQueue[1:len(client.RequestQueue)]
	return &rs
}

// ByAddress - get client by his IP and port addresses
func (cm *ClientMap) ByAddress(ip string, port uint16) (*Client, error) {
	var clientFound *Client
	clientFound = nil
	log.Printf("The client list = %v", cm.ClientList)
	for idx, client := range cm.ClientList {
		if client.IP == ip && client.Port == port {
			clientFound = &cm.ClientList[idx]
			break
		}
	}
	if clientFound == nil {
		return nil, errors.New(fmt.Sprintf("Could not find a client with address %v:%v", ip, port))
	}
	return clientFound, nil
}

// ByAddress - get client by his IP and port addresses
func (cm *ClientMap) ByUUID(uuid uuid.UUID) (*Client, error) {
	var clientFound *Client
	clientFound = nil
	log.Printf("The client list = %v", cm.ClientList)
	for idx, client := range cm.ClientList {
		if client.UUID == uuid {
			clientFound = &cm.ClientList[idx]
			break
		}
	}
	if clientFound == nil {
		return nil, errors.New(fmt.Sprintf("Could not find a client with UUID %v", uuid))
	}
	return clientFound, nil
}

func (cm *ClientMap) AddClient(c *Client) (*Client, error) {
	clientFound, err := cm.ByAddress(c.IP, c.Port)
	if err != nil {
		cm.ClientList = append(cm.ClientList, *c)
		c = &cm.ClientList[len(cm.ClientList)-1]
		return c, nil
	}
	return clientFound, err
}
