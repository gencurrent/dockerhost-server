/*
	The application essential variables
*/

package application

import (
	Types "../types"
)

var Config Types.Config
var ClientMap = Types.ClientMap{
	ClientList: []Types.Client{},
}