package http

import (
	"crypto/tls"
	"net/http"
	"sync"
	"time"
)

var (
	clientOnce sync.Once
	sharedClient *http.Client
	
	ollamaClientOnce sync.Once
	ollamaClient *http.Client
)

// createTransport creates a shared HTTP transport with optimized settings
func createTransport() *http.Transport {
	return &http.Transport{
		TLSHandshakeTimeout: 10 * time.Second,
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		DisableCompression:  true,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	}
}

// GetClient returns a shared HTTP client with optimized settings for cloud APIs
func GetClient() *http.Client {
	clientOnce.Do(func() {
		sharedClient = &http.Client{
			Timeout:   30 * time.Second,
			Transport: createTransport(),
		}
	})
	return sharedClient
}

// GetOllamaClient returns an HTTP client with extended timeout for local Ollama inference
func GetOllamaClient() *http.Client {
	ollamaClientOnce.Do(func() {
		ollamaClient = &http.Client{
			Timeout:   10 * time.Minute, // 10 minutes for local inference
			Transport: createTransport(),
		}
	})
	return ollamaClient
}