package rpc

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestGetParams(t *testing.T) {
	nodeURL := "https://rpc.tzbeta.net"
	apiTimeout := 10 * time.Second

	client, err := NewClient(nodeURL, &http.Client{
		Timeout: apiTimeout,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	ctx := context.Background()

	_, err = client.GetParams(ctx, BlockLevel(1))
	if err != nil {
		t.Errorf("client.GetParams: %v", err)
	}
}
