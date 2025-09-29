package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	inceptiondb "github.com/holaql/go-inceptiondb"
)

func main() {
	databaseID := valueOrEnv("INCEPTIONDB_DATABASE_ID", "")
	apiKey := valueOrEnv("INCEPTIONDB_API_KEY", "")
	apiSecret := valueOrEnv("INCEPTIONDB_API_SECRET", "")

	if databaseID == "" {
		log.Fatal("INCEPTIONDB_DATABASE_ID must be set")
	}
	if apiKey == "" || apiSecret == "" {
		log.Fatal("INCEPTIONDB_API_KEY and INCEPTIONDB_API_SECRET must be set")
	}

	cfg := inceptiondb.Config{
		DatabaseID: databaseID,
		APIKey:     apiKey,
		APISecret:  apiSecret,
	}
	if baseURL := strings.TrimSpace(os.Getenv("INCEPTIONDB_BASE_URL")); baseURL != "" {
		cfg.BaseURL = baseURL
	}

	client, err := inceptiondb.NewClient(cfg)
	if err != nil {
		log.Fatalf("create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collectionName := "todo-items"
	fmt.Printf("Using database %s, collection %s\n", databaseID, collectionName)

	_ = client.DropCollection(ctx, collectionName)

	if _, err := client.CreateCollection(ctx, &inceptiondb.CreateCollectionRequest{
		Name: collectionName,
		Defaults: map[string]any{
			"completed": false,
		},
	}); err != nil {
		log.Fatalf("create collection: %v", err)
	}

	todos := []map[string]any{
		{"id": "1", "title": "Write demo", "completed": false},
		{"id": "2", "title": "Test with real API", "completed": true},
		{"id": "3", "title": "Share results", "completed": false},
		{"id": "4", "title": "Record metrics", "completed": false},
		{"id": "5", "title": "Celebrate", "completed": false},
	}
	stream, err := client.InsertDocuments(ctx, collectionName, toAnySlice(todos)...)
	if err != nil {
		log.Fatalf("insert documents: %v", err)
	}
	fmt.Println("Inserted documents:")
	printStream(stream)

	findStream, err := client.Find(ctx, collectionName, &inceptiondb.FindRequest{
		QueryOptions: inceptiondb.QueryOptions{
			Filter: map[string]any{"completed": false},
			Limit:  2,
		},
	})
	if err != nil {
		log.Fatalf("find documents: %v", err)
	}
	fmt.Println("\nFirst page of incomplete tasks:")
	printStream(findStream)

	nextPage, err := client.Find(ctx, collectionName, &inceptiondb.FindRequest{
		QueryOptions: inceptiondb.QueryOptions{
			Filter: map[string]any{"completed": false},
			Skip:   2,
			Limit:  2,
		},
	})
	if err != nil {
		log.Fatalf("find documents: %v", err)
	}
	fmt.Println("\nSecond page of incomplete tasks:")
	printStream(nextPage)

	fmt.Println("\nCleaning up collection...")
	if err := client.DropCollection(ctx, collectionName); err != nil {
		log.Fatalf("drop collection: %v", err)
	}
}

func toAnySlice(items []map[string]any) []any {
	result := make([]any, len(items))
	for i, item := range items {
		result[i] = item
	}
	return result
}

func printStream(stream *inceptiondb.JSONStream) {
	defer stream.Close()

	err := inceptiondb.Iterate[map[string]any](stream, func(item *map[string]any) error {
		data, err := json.MarshalIndent(item, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	})
	if err != nil {
		log.Fatalf("stream error: %v", err)
	}
}

func valueOrEnv(envKey, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(envKey)); value != "" {
		return value
	}
	return fallback
}
