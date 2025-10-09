package http

import (
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestGetClient(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil client", func(t *testing.T) {
		t.Parallel()

		client := GetClient()
		if client == nil {
			t.Fatal("expected non-nil client")
		}
	})

	t.Run("returns same client on multiple calls", func(t *testing.T) {
		t.Parallel()

		client1 := GetClient()
		client2 := GetClient()

		if client1 != client2 {
			t.Fatal("expected same client instance on multiple calls")
		}
	})

	t.Run("client has correct timeout", func(t *testing.T) {
		t.Parallel()

		client := GetClient()
		expectedTimeout := 30 * time.Second

		if client.Timeout != expectedTimeout {
			t.Fatalf("expected timeout %v, got %v", expectedTimeout, client.Timeout)
		}
	})

	t.Run("client has custom transport", func(t *testing.T) {
		t.Parallel()

		client := GetClient()
		if client.Transport == nil {
			t.Fatal("expected client to have custom transport")
		}

		transport, ok := client.Transport.(*http.Transport)
		if !ok {
			t.Fatal("expected transport to be *http.Transport")
		}

		// Check some transport settings
		if transport.TLSHandshakeTimeout != 10*time.Second {
			t.Fatalf("expected TLS handshake timeout 10s, got %v", transport.TLSHandshakeTimeout)
		}

		if transport.MaxIdleConns != 10 {
			t.Fatalf("expected MaxIdleConns 10, got %d", transport.MaxIdleConns)
		}

		if transport.IdleConnTimeout != 30*time.Second {
			t.Fatalf("expected idle conn timeout 30s, got %v", transport.IdleConnTimeout)
		}

		if !transport.DisableCompression {
			t.Fatal("expected compression to be disabled")
		}

		if transport.TLSClientConfig == nil {
			t.Fatal("expected TLS client config to be set")
		}

		if transport.TLSClientConfig.InsecureSkipVerify != false {
			t.Fatal("expected InsecureSkipVerify to be false")
		}
	})
}

func TestGetOllamaClient(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil client", func(t *testing.T) {
		t.Parallel()

		client := GetOllamaClient()
		if client == nil {
			t.Fatal("expected non-nil client")
		}
	})

	t.Run("returns same client on multiple calls", func(t *testing.T) {
		t.Parallel()

		client1 := GetOllamaClient()
		client2 := GetOllamaClient()

		if client1 != client2 {
			t.Fatal("expected same client instance on multiple calls")
		}
	})

	t.Run("client has correct timeout", func(t *testing.T) {
		t.Parallel()

		client := GetOllamaClient()
		expectedTimeout := 10 * time.Minute

		if client.Timeout != expectedTimeout {
			t.Fatalf("expected timeout %v, got %v", expectedTimeout, client.Timeout)
		}
	})

	t.Run("client has custom transport", func(t *testing.T) {
		t.Parallel()

		client := GetOllamaClient()
		if client.Transport == nil {
			t.Fatal("expected client to have custom transport")
		}

		transport, ok := client.Transport.(*http.Transport)
		if !ok {
			t.Fatal("expected transport to be *http.Transport")
		}

		// Check some transport settings
		if transport.TLSHandshakeTimeout != 10*time.Second {
			t.Fatalf("expected TLS handshake timeout 10s, got %v", transport.TLSHandshakeTimeout)
		}

		if transport.MaxIdleConns != 10 {
			t.Fatalf("expected MaxIdleConns 10, got %d", transport.MaxIdleConns)
		}

		if transport.IdleConnTimeout != 30*time.Second {
			t.Fatalf("expected idle conn timeout 30s, got %v", transport.IdleConnTimeout)
		}

		if !transport.DisableCompression {
			t.Fatal("expected compression to be disabled")
		}

		if transport.TLSClientConfig == nil {
			t.Fatal("expected TLS client config to be set")
		}

		if transport.TLSClientConfig.InsecureSkipVerify != false {
			t.Fatal("expected InsecureSkipVerify to be false")
		}
	})
}

func TestClientsAreDifferent(t *testing.T) {
	t.Parallel()

	// Ensure that GetClient and GetOllamaClient return different instances
	regularClient := GetClient()
	ollamaClient := GetOllamaClient()

	if regularClient == ollamaClient {
		t.Fatal("expected regular client and ollama client to be different instances")
	}

	// Check that they have different timeouts
	if regularClient.Timeout == ollamaClient.Timeout {
		t.Fatal("expected regular client and ollama client to have different timeouts")
	}
}

func TestCreateTransport(t *testing.T) {
	t.Parallel()

	// We can't directly test createTransport since it's not exported,
	// but we can test its effects through GetClient
	client := GetClient()
	transport := client.Transport.(*http.Transport)

	// Test all the transport settings
	expectedSettings := map[string]interface{}{
		"TLSHandshakeTimeout": 10 * time.Second,
		"MaxIdleConns":        10,
		"IdleConnTimeout":     30 * time.Second,
		"DisableCompression":  true,
	}

	if transport.TLSHandshakeTimeout != expectedSettings["TLSHandshakeTimeout"] {
		t.Fatalf("expected TLSHandshakeTimeout %v, got %v",
			expectedSettings["TLSHandshakeTimeout"], transport.TLSHandshakeTimeout)
	}

	if transport.MaxIdleConns != expectedSettings["MaxIdleConns"] {
		t.Fatalf("expected MaxIdleConns %v, got %v",
			expectedSettings["MaxIdleConns"], transport.MaxIdleConns)
	}

	if transport.IdleConnTimeout != expectedSettings["IdleConnTimeout"] {
		t.Fatalf("expected IdleConnTimeout %v, got %v",
			expectedSettings["IdleConnTimeout"], transport.IdleConnTimeout)
	}

	if transport.DisableCompression != expectedSettings["DisableCompression"] {
		t.Fatalf("expected DisableCompression %v, got %v",
			expectedSettings["DisableCompression"], transport.DisableCompression)
	}

	// Test TLS config
	if transport.TLSClientConfig == nil {
		t.Fatal("expected TLS client config to be set")
	}

	if transport.TLSClientConfig.InsecureSkipVerify {
		t.Fatal("expected InsecureSkipVerify to be false")
	}
}

func TestConcurrentAccess(t *testing.T) {
	t.Parallel()

	t.Run("concurrent access to GetClient", func(t *testing.T) {
		t.Parallel()

		var wg sync.WaitGroup
		numGoroutines := 10
		clients := make([]*http.Client, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				clients[index] = GetClient()
			}(i)
		}

		wg.Wait()

		// All clients should be the same instance
		firstClient := clients[0]
		for i, client := range clients {
			if client != firstClient {
				t.Fatalf("client at index %d is different from first client", i)
			}
		}
	})

	t.Run("concurrent access to GetOllamaClient", func(t *testing.T) {
		t.Parallel()

		var wg sync.WaitGroup
		numGoroutines := 10
		clients := make([]*http.Client, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				clients[index] = GetOllamaClient()
			}(i)
		}

		wg.Wait()

		// All clients should be the same instance
		firstClient := clients[0]
		for i, client := range clients {
			if client != firstClient {
				t.Fatalf("client at index %d is different from first client", i)
			}
		}
	})

	t.Run("concurrent access to both clients", func(t *testing.T) {
		t.Parallel()

		var wg sync.WaitGroup
		numGoroutines := 20
		regularClients := make([]*http.Client, numGoroutines/2)
		ollamaClients := make([]*http.Client, numGoroutines/2)

		for i := 0; i < numGoroutines/2; i++ {
			wg.Add(2)
			go func(index int) {
				defer wg.Done()
				regularClients[index] = GetClient()
			}(i)
			go func(index int) {
				defer wg.Done()
				ollamaClients[index] = GetOllamaClient()
			}(i)
		}

		wg.Wait()

		// All regular clients should be the same instance
		firstRegularClient := regularClients[0]
		for i, client := range regularClients {
			if client != firstRegularClient {
				t.Fatalf("regular client at index %d is different from first regular client", i)
			}
		}

		// All ollama clients should be the same instance
		firstOllamaClient := ollamaClients[0]
		for i, client := range ollamaClients {
			if client != firstOllamaClient {
				t.Fatalf("ollama client at index %d is different from first ollama client", i)
			}
		}

		// Regular and ollama clients should be different
		if firstRegularClient == firstOllamaClient {
			t.Fatal("regular client and ollama client should be different instances")
		}
	})
}

func TestClientTimeoutValues(t *testing.T) {
	t.Parallel()

	regularClient := GetClient()
	ollamaClient := GetOllamaClient()

	// Regular client should have 30 second timeout
	expectedRegularTimeout := 30 * time.Second
	if regularClient.Timeout != expectedRegularTimeout {
		t.Fatalf("expected regular client timeout %v, got %v", expectedRegularTimeout, regularClient.Timeout)
	}

	// Ollama client should have 10 minute timeout
	expectedOllamaTimeout := 10 * time.Minute
	if ollamaClient.Timeout != expectedOllamaTimeout {
		t.Fatalf("expected ollama client timeout %v, got %v", expectedOllamaTimeout, ollamaClient.Timeout)
	}

	// Regular client timeout should be shorter than ollama client timeout
	if regularClient.Timeout >= ollamaClient.Timeout {
		t.Fatal("expected regular client timeout to be shorter than ollama client timeout")
	}
}
