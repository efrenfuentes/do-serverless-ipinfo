package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

type Request struct {
	IP string `json:"ip"`
}

type Response struct {
	StatusCode int               `json:"statusCode,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string            `json:"body,omitempty"`
}

var client = &http.Client{
	Timeout: 5 * time.Second,
	CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

const lookupError = `{"error":"No se pudo obtener la información del IP."}`

func failureStatus(err error) int {
	if errors.Is(err, context.DeadlineExceeded) {
		return http.StatusGatewayTimeout
	}
	return http.StatusBadGateway
}

func ipInfo(ip string) (int, string) {
	req, err := http.NewRequest(http.MethodGet, "https://ipinfo.io/"+ip+"/json", nil)
	if err != nil {
		return http.StatusBadGateway, lookupError
	}
	if token := os.Getenv("IPINFO_ACCESS_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := client.Do(req)
	if err != nil {
		return failureStatus(err), lookupError
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return resp.StatusCode, `{"error":"Se alcanzó el límite de solicitudes de IPinfo."}`
	}
	if resp.StatusCode != http.StatusOK {
		return http.StatusBadGateway, lookupError
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return failureStatus(err), lookupError
	}
	if !json.Valid(body) {
		return http.StatusBadGateway, lookupError
	}
	return http.StatusOK, string(body)
}

func Main(in Request) (*Response, error) {
	response := &Response{Headers: map[string]string{"Content-Type": "application/json"}}
	ip := strings.TrimSpace(in.IP)
	if net.ParseIP(ip) == nil {
		response.StatusCode = http.StatusBadRequest
		response.Body = `{"error":"Se requiere una dirección IP válida."}`
		return response, nil
	}

	response.StatusCode, response.Body = ipInfo(ip)
	return response, nil
}
