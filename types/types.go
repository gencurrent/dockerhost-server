package types

import (
	"encoding/json"
)

type RequestStructure struct {
	Request   string                 `json:"request"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ResponseStructure Структура ответа от сервера
type ResponseStructure struct {
	Request   string                 `json:"request"`
	Arguments map[string]interface{} `json:"arguments"`
}

type ResponseTemp struct {
	Response string
	Error    error
}

// marshal Encode to JSON String
func (r *RequestStructure) Marshal() (string, error) {
	result, err := json.Marshal(r)
	return string(result), err
}
