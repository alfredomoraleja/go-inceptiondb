package inceptiondb

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOnline(t *testing.T) {
	t.SkipNow()
	c, err := NewClient(Config{BaseURL: "https://inceptiondb.io"})
	if err != nil {
		t.Fatal(err)
	}
	cols, err := c.ListCollections(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(cols)
}

func TestEncodeQueryRequestNil(t *testing.T) {
	reader, err := encodeQueryRequest((*FindRequest)(nil))
	if err != nil {
		t.Fatalf("encodeQueryRequest() error = %v", err)
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if got := string(data); got != "{}" {
		t.Fatalf("encodeQueryRequest() = %s, want {}", got)
	}
}

func TestEncodeQueryRequestPayload(t *testing.T) {
	req := &FindRequest{
		QueryOptions: QueryOptions{
			Index:  "my-index",
			Limit:  5,
			Filter: map[string]any{"name": "Fulanez"},
		},
	}
	reader, err := encodeQueryRequest(req)
	if err != nil {
		t.Fatalf("encodeQueryRequest() error = %v", err)
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	payload := map[string]any{}
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if got := payload["index"]; got != req.Index {
		t.Fatalf("index = %v, want %s", got, req.Index)
	}
	if got := payload["limit"]; got != float64(req.Limit) {
		t.Fatalf("limit = %v, want %d", got, req.Limit)
	}
	if _, ok := payload["filter"].(map[string]any); !ok {
		t.Fatalf("filter type = %T, want map[string]any", payload["filter"])
	}
}

func TestClientAddsCredentialsAndDatabasePath(t *testing.T) {
	t.Parallel()

	var (
		receivedPath   string
		receivedAPIKey string
		receivedSecret string
		receivedMethod string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		receivedAPIKey = r.Header.Get("Api-Key")
		receivedSecret = r.Header.Get("Api-Secret")
		receivedMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `[]`)
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(Config{
		BaseURL:    server.URL,
		DatabaseID: "my-db",
		APIKey:     "my-key",
		APISecret:  "my-secret",
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	if _, err := client.ListCollections(context.Background()); err != nil {
		t.Fatalf("ListCollections() error = %v", err)
	}

	if receivedMethod != http.MethodGet {
		t.Fatalf("method = %s, want GET", receivedMethod)
	}
	if receivedPath != "/v1/databases/my-db/collections" {
		t.Fatalf("path = %s, want /v1/databases/my-db/collections", receivedPath)
	}
	if receivedAPIKey != "my-key" {
		t.Fatalf("Api-Key = %s, want my-key", receivedAPIKey)
	}
	if receivedSecret != "my-secret" {
		t.Fatalf("Api-Secret = %s, want my-secret", receivedSecret)
	}
}

func TestNewClientDefaultsBaseURL(t *testing.T) {
	t.Parallel()

	client, err := NewClient(Config{})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	if got, want := client.baseURL.String(), "https://inceptiondb.hola.cloud"; got != want {
		t.Fatalf("baseURL = %s, want %s", got, want)
	}
}
