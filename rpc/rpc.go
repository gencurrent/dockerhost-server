// /*
// 	The Remote Procedure Call package
// 	Drives all the transmission between the server and a client
// */
// package rpc

// import (
// 	Types "../types"

// 	RequestHandlers "../handlers"
// 	FBHandlers "../feedbackhandlers"
// )

// type DHRPCMethod struct {
// 	Name            string
// 	RequestHandler  *func() (string, error)
// 	FeedBackHandler *func(*Types.Client) (error)
// }

// var DHRPCMethodTable = []DHRPCMethod{
// 	DHRPCMethod{
// 		"Status",
// 		RequestHandlers.Status,


// 	}
// }