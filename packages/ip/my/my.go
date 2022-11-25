package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
)

// Response returns back the http code, type of data, and the ip to the user.
type Response struct {
	// StatusCode is the http code that will be returned back to the user.
	StatusCode int `json:"statusCode,omitempty"`
	// Headers is the information about the type of data being returned back.
	Headers map[string]string `json:"headers,omitempty"`
	// Body will contain the presigned url to upload or download files.
	Body string `json:"body,omitempty"`
}

type IPResponse struct {
	IP string `json:"ip"`
}

func (r IPResponse) ToString() string {
	bodyJson := ""
	body, err := json.Marshal(r)
	if err != nil {
		log.Println("No se pudo obtener la IP.")
		bodyJson = "{\"error\": \"No se pudo obtener la IP.\"}"
	} else {
		bodyJson = string(body)
	}

	return bodyJson
}

func ResponseHeaders() map[string]string {
	return map[string]string{
		"Content-Type": "application/json",
	}
}

func GetRequestIp(headers map[string]interface{}) string {
	ips := headers["x-forwarded-for"].(string)

	ip := strings.Split(ips, ",")[0]

	fmt.Println("IP: ", ip)

	return ip
}

func Main(args map[string]interface{}) (*Response, error) {
	if args["__ow_headers"] == nil {
		log.Println("No se encontraron headers.")
		return &Response{
			StatusCode: 400,
			Headers:    ResponseHeaders(),
			Body:       "{\"error\": \"No se pudo obtener la IP.\"}",
		}, errors.New("No se encontraron headers.")
	}

	headers := args["__ow_headers"].(map[string]interface{})

	ip := IPResponse{IP: GetRequestIp(headers)}

	return &Response{
		StatusCode: 200,
		Headers:    ResponseHeaders(),
		Body:       ip.ToString(),
	}, nil
}
