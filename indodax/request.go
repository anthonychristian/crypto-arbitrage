// Package indodax is a wrapper for interacting with indodax's API
// Since there is no websocket API for indodax,
// we will use gorequest to continuously grab indodax's depth
package indodax

import (
	"encoding/json"
	"fmt"

	"github.com/parnurzeal/gorequest"
)

// base URL for the api gateway
const (
	baseURL  = "https://indodax.com/api/"
	endpoint = "/depth"
)

// IndodaxAPI serves the app for interacting with HTTP endpoints
// req <- the request object
// add some more functionalities in the future(maybe retries,
// error handling, etc)
type IndodaxAPI struct {
	req *gorequest.SuperAgent
}

var IndodaxInstance *IndodaxAPI

func InitIndodax() *IndodaxAPI {
	IndodaxInstance = &IndodaxAPI{
		req: gorequest.New(),
	}
	return IndodaxInstance
}

func (i *IndodaxAPI) getDepth(symbol string) (dat Depth) {
	// check if symbol is valid
	// build request
	req, body, errs := i.req.Get(baseURL + symbol + endpoint).
		End()
	if errs != nil || req.StatusCode != 200 {
		//error handling here
		return dat
	}
	err := json.Unmarshal([]byte(body), &dat)
	if err != nil {
		fmt.Println("error when unmarshalling the body")
		return dat
	}
	return dat
}
