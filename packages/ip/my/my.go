package main

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
)

type Response struct {
	StatusCode int               `json:"statusCode,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string            `json:"body,omitempty"`
}

type IPResponse struct {
	IP string `json:"ip"`
}

func Main(args map[string]interface{}) (*Response, error) {
	var headers map[string]interface{}
	if event, present := args["http"]; present {
		httpEvent, _ := event.(map[string]interface{})
		headers, _ = httpEvent["headers"].(map[string]interface{})
	} else {
		headers, _ = args["__ow_headers"].(map[string]interface{})
	}
	forwardedFor, _ := headers["x-forwarded-for"].(string)
	first := strings.SplitN(forwardedFor, ",", 2)[0]
	ip := strings.TrimSpace(first)
	response := &Response{Headers: map[string]string{"Content-Type": "application/json"}}
	if net.ParseIP(ip) == nil {
		response.StatusCode = http.StatusBadRequest
		response.Body = `{"error":"No se pudo obtener la IP."}`
		return response, nil
	}

	body, err := json.Marshal(IPResponse{IP: ip})
	if err != nil {
		return nil, err
	}
	response.StatusCode = http.StatusOK
	response.Body = string(body)
	return response, nil
}
