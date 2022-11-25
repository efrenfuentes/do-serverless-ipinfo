package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// Request takes a ip to returns the ip info.
type Request struct {
	IP string `json:"ip"`
}

// Response returns back the http code, type of data, and the ip info to the user.
type Response struct {
	// StatusCode is the http code that will be returned back to the user.
	StatusCode int `json:"statusCode,omitempty"`
	// Headers is the information about the type of data being returned back.
	Headers map[string]string `json:"headers,omitempty"`
	// Body will contain the ip info.
	Body string `json:"body,omitempty"`
}

func ResponseHeaders() map[string]string {
	return map[string]string{
		"Content-Type": "application/json",
	}
}

func ipInfo(ip string) string {
	accessToken := os.Getenv("IPINFO_ACCESS_TOKEN")
	fmt.Println("IPINFO_ACCESS_TOKEN: ", accessToken)

	url := fmt.Sprintf("http://ipinfo.io/%s?token=%s", ip, accessToken)

	// Make the request.
	client := http.DefaultClient
	resp, err := client.Get(url)
	if err != nil {
		return "{\"error\": \"No se pudo obtener la información del IP.\"}"
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "{\"error\": \"No se pudo obtener la información del IP.\"}"
	}

	info := string(b)

	return info

}

func Main(in Request) (*Response, error) {
	if in.IP == "" {
		log.Println("No se encontró la IP.")
		return &Response{
			StatusCode: 400,
			Headers:    ResponseHeaders(),
			Body:       "{\"error\": \"No se pudo obtener la IP.\"}",
		}, nil
	}

	return &Response{
		StatusCode: 200,
		Headers:    ResponseHeaders(),
		Body:       ipInfo(in.IP),
	}, nil
}
