package client_test

import (
	"context"
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
	tests := []struct {
		name         string
		mockResponse []byte
		wantErr      error
	}{
		{
			name:         "success",
			mockResponse: mustMockResponse(t, "testdata/response.json"),
			wantErr:      nil,
		},
		{
			name:         "finish reason length, no tool call",
			mockResponse: mustMockResponse(t, "testdata/response_length_text.json"),
			wantErr:      client.ErrTruncated,
		},
		{
			name:         "finish reason length, has tool call",
			mockResponse: mustMockResponse(t, "testdata/response_length_tool_call.json"),
			wantErr:      client.ErrTruncated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, err := w.Write(tt.mockResponse)
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
			ctx := context.Background()
			got, err := c.Request(ctx, "test prompt")

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)

				return
			}

			require.NoError(t, err)
			require.Equal(t, want, got)
		})
	}
}

func mustMockResponse(t *testing.T, path string) []byte {
	t.Helper()

	mockResponse, err := os.ReadFile(path)
	require.NoError(t, err)

	return mockResponse
}
