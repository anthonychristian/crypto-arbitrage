// Package indodax is a wrapper for interacting with indodax's API
// Since there is no websocket API for indodax,
// we will use gorequest to continuously grab indodax's depth
package indodax

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/parnurzeal/gorequest"
)

// base URL for the api gateway
const (
	publicBaseURL = "https://indodax.com/api/"
	depthEndpoint = "/depth"

	APIBaseURL = "https://indodax.com/tapi"
)

type Balance struct {
	Map map[string]float64
}

type IdxResponse struct {
	Success int         `json:"success"`
	Return  interface{} `json:"return"`
	Error   interface{} `json:"error"`
}

// IndodaxAPI serves the app for interacting with HTTP endpoints
// req <- the request object
// add some more functionalities in the future(maybe retries,
// error handling, etc)
type IndodaxAPI struct {
	req       *gorequest.SuperAgent
	apiKey    string
	secretKey string
}

var IndodaxInstance *IndodaxAPI

func InitIndodax(api, secret string) *IndodaxAPI {
	IndodaxInstance = &IndodaxAPI{
		req:       gorequest.New(),
		apiKey:    api,
		secretKey: secret,
	}
	return IndodaxInstance
}

func (i *IndodaxAPI) GetDepth(symbol string) (dat Depth) {
	// check if symbol is valid
	// build request
	resp, body, errs := i.req.Get(publicBaseURL + symbol + depthEndpoint).
		End()
	if resp == nil || resp.StatusCode != 200 || errs != nil {
		//error handling here
		// log.Fatal(dat, errs[0].Error())
		// return dat
		return Depth{}
	}
	err := json.Unmarshal([]byte(body), &dat)
	if err != nil {
		fmt.Println("error when unmarshalling the body")
		return dat
	}
	return dat
}

func (i *IndodaxAPI) GetInfo() interface{} {
	timeStr := strconv.FormatInt(time.Now().Unix(), 10)
	params := "method=getInfo&nonce=" + timeStr
	// build request
	resp, body, errs := i.req.
		Post(APIBaseURL).
		Set("Key", i.apiKey).
		Set("Sign", sign(params, i.secretKey)).
		Send(params).
		End()
	if resp.StatusCode != 200 || errs != nil {
		fmt.Println(errs)
		//error handling here
		return body
	}

	return body
}
