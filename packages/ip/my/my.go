package main

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

type Response struct {
	StatusCode int               `json:"statusCode,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string            `json:"body,omitempty"`
}

type IPResponse struct {
	IP     string          `json:"ip"`
	Family string          `json:"family"`
	Info   json.RawMessage `json:"info"`
	Error  string          `json:"error,omitempty"`
}

var client = &http.Client{
	Timeout: 5 * time.Second,
	CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

const lookupError = "No se pudo obtener la información del IP."

func metadata(ip string) (json.RawMessage, string) {
	req, err := http.NewRequest(http.MethodGet, "https://ipinfo.io/"+ip+"/json", nil)
	if err != nil {
		return nil, lookupError
	}
	if token := os.Getenv("IPINFO_ACCESS_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, lookupError
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, "Se alcanzó el límite de solicitudes de IPinfo."
	}
	if resp.StatusCode != http.StatusOK {
		return nil, lookupError
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil || !json.Valid(body) {
		return nil, lookupError
	}
	return json.RawMessage(body), ""
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
	addr := net.ParseIP(ip)
	if addr == nil {
		response.StatusCode = http.StatusBadRequest
		response.Body = `{"error":"No se pudo obtener la IP."}`
		return response, nil
	}

	result := IPResponse{IP: ip, Family: "ipv6"}
	if addr.To4() != nil {
		result.Family = "ipv4"
	}
	result.Info, result.Error = metadata(ip)
	body, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	response.StatusCode = http.StatusOK
	response.Body = string(body)
	return response, nil
}
