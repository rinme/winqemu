package msdl

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetProducts(t *testing.T) {
	mockJSON := `{
		"3813": { "name": "Windows 11 26H2 (26300.9457)", "badge": "LATEST", "archs": ["x64"] },
		"2618": { "name": "Windows 10 22H2 v1 (19045.2965)", "badge": "EOL", "archs": ["x64"] }
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockJSON))
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	products, err := client.GetProducts(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(products) != 2 {
		t.Errorf("expected 2 products, got %d", len(products))
	}
}

func TestResolveConsumerLink(t *testing.T) {
	skuJSON := `{"Skus":[{"Id":"27160","Language":"English","ProductDisplayName":"Windows 11 Client"}]}`
	proxyJSON := `{"ProductDownloadOptions":[{"Name":"Win11 English","Uri":"https://cdn.microsoft.com/win11-26h2.iso","DownloadType":1}]}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/skuinfo" {
			w.Write([]byte(skuJSON))
			return
		}
		if r.URL.Path == "/proxy" {
			w.Write([]byte(proxyJSON))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	url, err := client.ResolveConsumerLink(context.Background(), "3813", "English", "x64")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "https://cdn.microsoft.com/win11-26h2.iso"
	if url != expected {
		t.Errorf("expected %s, got %s", expected, url)
	}
}

func TestResolveEvalLink(t *testing.T) {
	evalJSON := `{"product":"win11-ent","links":[{"arch":"x64","lang":"en-us","url":"https://cdn.microsoft.com/win11-ent.iso"}]}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/evallinks" {
			w.Write([]byte(evalJSON))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	url, err := client.ResolveEvalLink(context.Background(), "win11-ent", "en-us", "x64")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "https://cdn.microsoft.com/win11-ent.iso"
	if url != expected {
		t.Errorf("expected %s, got %s", expected, url)
	}
}
