// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"internal/clilog"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

// RateLimitedHttpClient
type RateLimitedHTTPClient struct {
	client      *http.Client
	Ratelimiter *rate.Limiter
}

// allow 6 every 1 second (360 per min, limit is 480 per min)
var integrationAPIRateLimit = rate.NewLimiter(rate.Every(time.Second), 6)

// allow 1 every 1 second (60 per min, limit is 120 per min)
var connectorsAPIRateLimit = rate.NewLimiter(rate.Every(time.Second), 1)

// disable rate limit
var noAPIRateLimit = rate.NewLimiter(rate.Inf, 1)

// HttpClient method is used to GET,POST,PUT or DELETE JSON data
func HttpClient(params ...string) APIResponse {
	// The first parameter is url. If only one parameter is sent, assume GET
	// The second parameter is the payload. The two parameters are sent, assume POST
	// THe third parameter is the method. If three parameters are sent, assume method in param
	// The fourth parameter is content type
	var req *http.Request
	contentType := "application/json"

	client, err := getHttpClient()
	if err != nil {
		return APIResponse{
			RespBody: nil,
			Err:      err,
		}
	}

	clilog.Debug.Println("Connecting to: ", params[0])
	ctx := context.Background()

	switch paramLen := len(params); paramLen {
	case 1:
		clilog.Debug.Println("Method: GET")
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, params[0], nil)
	case 2:
		// some POST functions don't have a body
		clilog.Debug.Println("Method: POST")
		if len([]byte(params[1])) > 0 {
			payload, _ := PrettifyJson([]byte(params[1]))
			clilog.Debug.Println("Payload: ", string(payload))
		}
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, params[0], bytes.NewBuffer([]byte(params[1])))
	case 3:
		if req, err = getRequest(params); err != nil {
			return APIResponse{
				RespBody: nil,
				Err:      err,
			}
		}
	case 4:
		if req, err = getRequest(params); err != nil {
			return APIResponse{
				RespBody: nil,
				Err:      err,
			}
		}
		contentType = params[3]
	default:
		return APIResponse{
			RespBody: nil,
			Err:      NewCliError("unsupported method", nil),
		}
	}

	if err != nil {
		clilog.Debug.Println("error in client: ", err)
		return APIResponse{
			RespBody: nil,
			Err:      NewCliError("error in client", err),
		}
	}

	req, err = setAuthHeader(req)
	if err != nil {
		return APIResponse{
			RespBody: nil,
			Err:      NewCliError("error setting auth header", err),
		}
	}

	clilog.Debug.Println("Content-Type : ", contentType)
	req.Header.Set("Content-Type", contentType)

	if DryRun() {
		return APIResponse{
			RespBody: nil,
			Err:      nil,
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		clilog.Debug.Println("error connecting: ", err)
		return APIResponse{
			RespBody: nil,
			Err:      NewCliError("error connecting", err),
		}
	}

	return handleResponse(resp)
}

// PrettyPrint method prints formatted json
func PrettyPrint(response APIResponse) error {
	if response.Err != nil {
		return response.Err
	}
	if response.RespBody == nil || len(response.RespBody) == 0 || strings.TrimSpace(string(response.RespBody)) == "{}" {
		clilog.Info.Println("Action completed successfully")
		return nil
	}

	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, response.RespBody, "", "\t")
	if err != nil {
		clilog.Debug.Println("error parsing response: ", err)
		return NewCliError("error parsing response", err)
	}

	clilog.HTTPResponse.Println(prettyJSON.String())
	return nil
}

func PrettifyJson(body []byte) (prettyJson []byte, err error) {
	prettyJSON := bytes.Buffer{}
	err = json.Indent(&prettyJSON, body, "", "\t")
	if err != nil {
		clilog.Debug.Printf("error parsing json response: %v, the original response was: %s\n", err, string(body))
		return nil, NewCliError("error parsing json response", err)
	}
	return prettyJSON.Bytes(), nil
}

func getRequest(params []string) (req *http.Request, err error) {
	ctx := context.Background()
	if params[2] == "DELETE" {
		clilog.Debug.Println("Method: DELETE")
		req, err = http.NewRequestWithContext(ctx, http.MethodDelete, params[0], nil)
	} else if params[2] == "PUT" {
		clilog.Debug.Println("Method: PUT")
		clilog.Debug.Println("Payload: ", params[1])
		req, err = http.NewRequestWithContext(ctx, http.MethodPut, params[0], bytes.NewBuffer([]byte(params[1])))
	} else if params[2] == "PATCH" {
		clilog.Debug.Println("Method: PATCH")
		clilog.Debug.Println("Payload: ", params[1])
		req, err = http.NewRequestWithContext(ctx, http.MethodPatch, params[0], bytes.NewBuffer([]byte(params[1])))
	} else if params[2] == "POST" {
		clilog.Debug.Println("Method: POST")
		clilog.Debug.Println("Payload: ", params[1])
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, params[0], bytes.NewBuffer([]byte(params[1])))
	} else {
		return nil, NewCliError("unsupported method", nil)
	}
	return req, nil
}

func setAuthHeader(req *http.Request) (*http.Request, error) {
	if GetIntegrationToken() == "" {
		if err := SetAccessToken(); err != nil {
			return nil, NewCliError("error setting token", err)
		}
	}
	clilog.Debug.Println("Setting token : ", GetIntegrationToken())
	req.Header.Add("Authorization", "Bearer "+GetIntegrationToken())
	return req, nil
}

// Do the HTTP request
func (c *RateLimitedHTTPClient) Do(req *http.Request) (*http.Response, error) {
	ctx := context.Background()
	// Wait until the rate is below Apigee limits
	err := c.Ratelimiter.Wait(ctx)
	if err != nil {
		return nil, NewCliError("unable to create a rate limiter", err)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, NewCliError("unable to create client", err)
	}
	return resp, nil
}

func getHttpClient() (client *RateLimitedHTTPClient, err error) {
	var apiRateLimit *rate.Limiter

	switch r := GetRate(); r {
	case IntegrationAPI:
		apiRateLimit = integrationAPIRateLimit
	case ConnectorsAPI:
		apiRateLimit = connectorsAPIRateLimit
	case None:
		apiRateLimit = noAPIRateLimit
	default:
		apiRateLimit = noAPIRateLimit
	}

	if GetProxyURL() != "" {
		var proxyUrl *url.URL
		if proxyUrl, err = url.Parse(GetProxyURL()); err != nil {
			return nil, NewCliError("error parsing proxy url", err)
		}
		return &RateLimitedHTTPClient{
			client: &http.Client{
				Transport: &http.Transport{
					Proxy: http.ProxyURL(proxyUrl),
				},
			},
			Ratelimiter: apiRateLimit,
		}, nil
	} else {
		return &RateLimitedHTTPClient{
			client:      http.DefaultClient,
			Ratelimiter: apiRateLimit,
		}, nil
	}
}

func handleResponse(resp *http.Response) APIResponse {
	var respBody []byte
	var err error

	if resp != nil {
		defer resp.Body.Close()
	}

	if resp == nil {
		clilog.Debug.Println("error in response: Response was null")
		return APIResponse{
			RespBody: nil,
			Err:      NewCliError("error in response: Response was null", nil),
		}
	}

	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		clilog.Debug.Printf("error in response: %v\n", err)
		return APIResponse{
			RespBody: nil,
			Err:      NewCliError("error in response", err),
		}
	} else if resp.StatusCode > 399 {
		if GetConflictsAsErrors() && resp.StatusCode == http.StatusConflict {
			clilog.Warning.Printf("entity already exists, ignoring conflict")
			return APIResponse{
				RespBody: respBody,
				Err:      nil,
			}
		}
		clilog.Debug.Printf("status code %d, error in response: %s\n", resp.StatusCode, string(respBody))
		clilog.HTTPError.Println(string(respBody))
		return APIResponse{
			RespBody: nil,
			Err:      NewCliError(getErrorMessage(resp.StatusCode)+": "+string(respBody), nil),
		}
	}
	clilog.Debug.Println(string(respBody))
	return APIResponse{
		RespBody: respBody,
		Err:      nil,
	}
}

func getErrorMessage(statusCode int) string {
	switch statusCode {
	case 400:
		return "Bad Request - malformed request syntax"
	case 401:
		return "Unauthorized - the client must authenticate itself"
	case 403:
		return "Forbidden - the client does not have access rights"
	case 404:
		return "Not found - the server cannot find the requested resource"
	case 405:
		return "Method Not Allowed - the request method is not supported by the target resource"
	case 409:
		return "Conflict - request conflicts with the current state of the server"
	case 415:
		return "Unsupported media type - media format of the requested data is not supported by the server"
	case 429:
		return "Too Many Request - user has sent too many requests"
	case 500:
		return "Internal server error"
	case 501:
		return "Not Implemented - request method is not supported by the server"
	case 502:
		return "Bad Gateway"
	case 503:
		return "Service Unavaliable - the server is not ready to handle the request"
	default:
		return "unknown error"
	}
}
