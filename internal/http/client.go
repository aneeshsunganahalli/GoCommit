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

// GetClient returns a shared HTTP client with optimized settings for cloud APIs
func GetClient() *http.Client {
	clientOnce.Do(func() {
		transport := &http.Transport{
			TLSHandshakeTimeout: 10 * time.Second,
			MaxIdleConns:        10,
			IdleConnTimeout:     30 * time.Second,
			DisableCompression:  true,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
		}

		sharedClient = &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		}
	})
	return sharedClient
}

// GetOllamaClient returns an HTTP client with extended timeout for local Ollama inference
func GetOllamaClient() *http.Client {
	ollamaClientOnce.Do(func() {
		transport := &http.Transport{
			TLSHandshakeTimeout: 10 * time.Second,
			MaxIdleConns:        10,
			IdleConnTimeout:     30 * time.Second,
			DisableCompression:  true,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
		}

		ollamaClient = &http.Client{
			Timeout:   10 * time.Minute, // 10 minutes for local inference
			Transport: transport,
		}
	})
	return ollamaClient
}