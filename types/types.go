package types

type RequestStructure struct {
	Request   string                 `json:"request"`
	Arguments map[string]interface{} `json:"arguments"`
}

type ResponseStructure struct {
	Request   string                 `json:"request"`
	Arguments map[string]interface{} `json:"arguments"`
}

type ResponseTemp struct {
	Response string
	Error error
}