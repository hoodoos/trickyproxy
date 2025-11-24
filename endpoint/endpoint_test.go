package endpoint

import (
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestHTTPKeepAlive(t *testing.T) {
	var connCount int32
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Wrap the listener to count connections
	server.Listener = &connCounterListener{
		Listener:  server.Listener,
		connCount: &connCount,
	}

	server.Start()
	defer server.Close()

	// Extract host and port from the server URL
	host, port, err := net.SplitHostPort(server.URL[len("http://"):])
	if err != nil {
		t.Fatalf("Failed to parse server URL: %v", err)
	}

	// Create a new endpoint instance
	ep := New(host, port, "http", "", nil, nil, nil)

	// Make multiple requests
	numRequests := 5
	for i := 0; i < numRequests; i++ {
		req, _ := http.NewRequest("GET", server.URL, nil)
		_, _, err := ep.Do(req)
		if err != nil {
			t.Fatalf("Request %d failed: %v", i+1, err)
		}
	}

	// Check the connection count
	finalConnCount := atomic.LoadInt32(&connCount)
	if finalConnCount != 1 {
		t.Errorf("Expected 1 connection for %d requests, but got %d", numRequests, finalConnCount)
	}
}

func TestHTTPSKeepAlive(t *testing.T) {
	var connCount int32
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Wrap the listener to count connections
	server.Listener = &connCounterListener{
		Listener:  server.Listener,
		connCount: &connCount,
	}

	server.StartTLS()
	defer server.Close()

	// Extract host and port from the server URL
	host, port, err := net.SplitHostPort(server.URL[len("https://"):])
	if err != nil {
		t.Fatalf("Failed to parse server URL: %v", err)
	}

	// Create a new endpoint instance for TLS, without client certs
	ep := NewTLS("https", host, port, "", "", "")

	// Make multiple requests
	numRequests := 5
	for i := 0; i < numRequests; i++ {
		req, _ := http.NewRequest("GET", server.URL, nil)
		_, _, err := ep.Do(req)
		if err != nil {
			t.Fatalf("Request %d failed: %v", i+1, err)
		}
	}

	// Check the connection count
	finalConnCount := atomic.LoadInt32(&connCount)
	if finalConnCount != 1 {
		t.Errorf("Expected 1 connection for %d requests, but got %d", numRequests, finalConnCount)
	}
}

// connCounterListener wraps a net.Listener and counts accepted connections.
type connCounterListener struct {
	net.Listener
	connCount *int32
}

func (l *connCounterListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err == nil {
		atomic.AddInt32(l.connCount, 1)
	}
	return conn, err
}
