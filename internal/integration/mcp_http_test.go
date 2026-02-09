//go:build integration

package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestIntegration_MCP_HealthEndpoint(t *testing.T) {
	env := newTestEnv(t, nil)

	// Find a free port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := listener.Addr().String()
	listener.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- env.MCPServer.ServeHTTP(ctx, addr, nil)
	}()

	// Wait for server to start
	var resp *http.Response
	client := &http.Client{Timeout: 2 * time.Second}
	for i := 0; i < 20; i++ {
		resp, err = client.Get(fmt.Sprintf("http://%s/health", addr))
		if err == nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("health endpoint unreachable after retries: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	assertNoError(t, err)

	var healthResp struct {
		Status  string `json:"status"`
		Server  string `json:"server"`
		Version string `json:"version"`
	}
	if err := json.Unmarshal(body, &healthResp); err != nil {
		t.Fatalf("unmarshal health response: %v", err)
	}

	if healthResp.Status != "ok" {
		t.Errorf("expected status ok, got %s", healthResp.Status)
	}
	if healthResp.Server != "acai" {
		t.Errorf("expected server acai, got %s", healthResp.Server)
	}
	if healthResp.Version != "test" {
		t.Errorf("expected version test, got %s", healthResp.Version)
	}

	// Shutdown
	cancel()

	select {
	case srvErr := <-errCh:
		if srvErr != nil {
			t.Errorf("server error on shutdown: %v", srvErr)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("server did not shut down within timeout")
	}
}

func TestIntegration_MCP_ServeHTTP_Shutdown(t *testing.T) {
	env := newTestEnv(t, nil)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := listener.Addr().String()
	listener.Close()

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		errCh <- env.MCPServer.ServeHTTP(ctx, addr, nil)
	}()

	// Wait for server to be ready
	client := &http.Client{Timeout: 2 * time.Second}
	for i := 0; i < 20; i++ {
		resp, err := client.Get(fmt.Sprintf("http://%s/health", addr))
		if err == nil {
			resp.Body.Close()
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	// Cancel context to trigger graceful shutdown
	cancel()

	select {
	case srvErr := <-errCh:
		// Graceful shutdown should complete without error
		if srvErr != nil {
			t.Errorf("expected clean shutdown, got error: %v", srvErr)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("graceful shutdown did not complete within timeout")
	}
}

func TestIntegration_MCP_HealthEndpoint_WithExtraRoutes(t *testing.T) {
	env := newTestEnv(t, nil)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := listener.Addr().String()
	listener.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		env.MCPServer.ServeHTTP(ctx, addr, func(mux *http.ServeMux) {
			mux.HandleFunc("/custom", func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("custom-route"))
			})
		})
	}()

	client := &http.Client{Timeout: 2 * time.Second}
	var resp *http.Response
	for i := 0; i < 20; i++ {
		resp, err = client.Get(fmt.Sprintf("http://%s/custom", addr))
		if err == nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("custom endpoint unreachable: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "custom-route" {
		t.Errorf("expected 'custom-route', got %s", string(body))
	}
}
