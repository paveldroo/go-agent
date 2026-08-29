package tool_call

import (
	"encoding/json"
	"errors"
	"fmt"
)

var errBadArguments = errors.New("model produced invalid tool arguments")

type Function struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ToolCall struct {
	ID       string   `json:"id"`
	Index    int      `json:"index"`
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

func (tc ToolCall) Args(v any) error {
	err := json.Unmarshal([]byte(tc.Function.Arguments), v)
	if err != nil {
		return fmt.Errorf("%w: %s: %w", errBadArguments, tc.Function.Arguments, err)
	}

	return nil
}
