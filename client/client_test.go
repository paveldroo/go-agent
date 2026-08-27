package client_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/paveldroo/go-agent/client"
	"github.com/paveldroo/go-agent/config"
	"github.com/stretchr/testify/require"
)

func TestClient_Request(t *testing.T) {
	t.Parallel()
	mockResponse, err := os.ReadFile("testdata/response.json")
	require.NoError(t, err)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write(mockResponse)
		if err != nil {
			t.Fatalf("handler write to response: %s", err.Error())
		}
	}))
	defer mockServer.Close()

	want := "Why don't scientists trust atoms?\n\nBecause they **make up everything**! 😄"

	cfg := config.Config{
		APIKey:    "",
		LLMURL:    mockServer.URL,
		ModelName: "",
	}
	c := client.New(&cfg)
	got, err := c.Request("test prompt")

	require.NoError(t, err)
	require.Equal(t, want, got)
}
