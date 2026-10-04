package conversation_test

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/paveldroo/go-agent/client"
	"github.com/paveldroo/go-agent/conversation"
	"github.com/paveldroo/go-agent/conversation/mocks"
	"github.com/paveldroo/go-agent/tool/tool"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestConversation_Run(t *testing.T) {
	t.Parallel()

	mockLLMClient := mocks.NewLLMClient(t)
	mockLLMClient.EXPECT().Request(mock.Anything, mock.Anything).Return(buildChatResponse(t, "testdata/response_tool_call.json"), nil).Once()
	mockLLMClient.EXPECT().Tools().Return([]tool.Tool{tool.WeatherTool()}).Once()
	mockLLMClient.EXPECT().Request(mock.Anything, mock.Anything).Return(buildChatResponse(t, "testdata/response_final.json"), nil).Once()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	conv := conversation.New()
	prompt := "what is the weather in Paris?"
	err = conv.Run(mockLLMClient, prompt)
	require.NoError(t, err)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	written, err := io.Copy(&buf, r)
	require.NoError(t, err)
	require.Positive(t, written)

	require.Equal(t, "The weather in Paris is -99°C, raining frogs.\n", buf.String())
	require.Len(t, conv.History, 4)

	llmMessage := conv.History[1]
	require.NotEmpty(t, llmMessage.ToolCalls)
	toolCall := llmMessage.ToolCalls[0]
	require.Equal(t, "get_weather", toolCall.Function.Name)

	toolCallMessage := conv.History[2]
	require.Equal(t, "-99°C, raining frogs in Paris", toolCallMessage.Content)
}

func buildChatResponse(t *testing.T, fName string) client.ChatResponse {
	t.Helper()

	body := mustMockResponse(t, fName)

	var cr client.ChatResponse
	err := json.Unmarshal(body, &cr)
	require.NoError(t, err)

	return cr
}

func mustMockResponse(t *testing.T, path string) []byte {
	t.Helper()

	mockResponse, err := os.ReadFile(path)
	require.NoError(t, err)

	return mockResponse
}
