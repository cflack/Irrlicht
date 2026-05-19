package vibe

import (
	"testing"

	"irrlicht/core/pkg/tailer"
)

func TestParser_ParseLine(t *testing.T) {
	p := &Parser{}

	tests := []struct {
		name     string
		raw      map[string]interface{}
		wantType string // EventType to expect
		wantSkip bool
	}{
		{
			name: "assistant message with content",
			raw: map[string]interface{}{
				"role":    "assistant",
				"content": "Hello, I am Vibe!",
			},
			wantType: "assistant",
			wantSkip: false,
		},
		{
			name: "user message",
			raw: map[string]interface{}{
				"role":    "user",
				"content": "What is 2+2?",
			},
			wantType: "user",
			wantSkip: false,
		},
		{
			name: "system message",
			raw: map[string]interface{}{
				"role":    "system",
				"content": "You are a helpful assistant.",
			},
			wantType: "",
			wantSkip: true,
		},
		{
			name: "tool result",
			raw: map[string]interface{}{
				"role":         "tool",
				"name":         "bash",
				"tool_call_id":  "call_123",
				"content":      "Command executed successfully",
			},
			wantType: "tool_result",
			wantSkip: false,
		},
		{
			name: "assistant with tool calls",
			raw: map[string]interface{}{
				"role": "assistant",
				"content": "Let me run a command",
				"tool_calls": []interface{}{
					map[string]interface{}{
						"id": "call_456",
						"function": map[string]interface{}{
							"name": "bash",
							"arguments": "ls -la",
						},
						"type": "function",
					},
				},
			},
			wantType: "assistant",
			wantSkip: false,
		},
		{
			name: "assistant with reasoning",
			raw: map[string]interface{}{
				"role":             "assistant",
				"content":          "The answer is 4",
				"reasoning_content": "Let me think about this... 2+2=4",
			},
			wantType: "assistant",
			wantSkip: false,
		},
		{
			name: "message with usage",
			raw: map[string]interface{}{
				"role":    "assistant",
				"content": "Response",
				"usage": map[string]interface{}{
					"prompt_tokens":     float64(100),
					"completion_tokens": float64(50),
				},
			},
			wantType: "assistant",
			wantSkip: false,
		},
		{
			name: "tool with error",
			raw: map[string]interface{}{
				"role":        "tool",
				"name":        "bash",
				"content":     "Error: command not found",
				"is_error":    true,
				"tool_call_id": "call_789",
			},
			wantType: "tool_result",
			wantSkip: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := p.ParseLine(tt.raw)

			if ev == nil {
				t.Fatal("ParseLine returned nil")
			}

			if ev.Skip != tt.wantSkip {
				t.Errorf("Skip = %v, want %v", ev.Skip, tt.wantSkip)
			}

			if ev.EventType != tt.wantType {
				t.Errorf("EventType = %q, want %q", ev.EventType, tt.wantType)
			}
		})
	}
}

func TestExtractVibeUsage(t *testing.T) {
	tests := []struct {
		name string
		usage map[string]interface{}
		want  *tailer.TokenSnapshot
	}{
		{
			name: "prompt and completion tokens",
			usage: map[string]interface{}{
				"prompt_tokens":     float64(100),
				"completion_tokens": float64(50),
			},
			want: &tailer.TokenSnapshot{Input: 100, Output: 50},
		},
		{
			name: "only prompt tokens",
			usage: map[string]interface{}{
				"prompt_tokens": float64(200),
			},
			want: &tailer.TokenSnapshot{Input: 200},
		},
		{
			name: "only completion tokens",
			usage: map[string]interface{}{
				"completion_tokens": float64(75),
			},
			want: &tailer.TokenSnapshot{Output: 75},
		},
		{
			name: "total tokens only",
			usage: map[string]interface{}{
				"total_tokens": float64(150),
			},
			want: &tailer.TokenSnapshot{Total: 150},
		},
		{
			name: "empty usage",
			usage: map[string]interface{}{},
			want:  nil,
		},
		{
			name: "nil usage",
			usage: nil,
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractVibeUsage(tt.usage)

			if tt.want == nil {
				if got != nil {
					t.Errorf("extractVibeUsage() = %+v, want nil", got)
				}
				return
			}

			if got == nil {
				t.Errorf("extractVibeUsage() = nil, want %+v", tt.want)
				return
			}

			if got.Input != tt.want.Input {
				t.Errorf("Input = %v, want %v", got.Input, tt.want.Input)
			}
			if got.Output != tt.want.Output {
				t.Errorf("Output = %v, want %v", got.Output, tt.want.Output)
			}
			if got.Total != tt.want.Total {
				t.Errorf("Total = %v, want %v", got.Total, tt.want.Total)
			}
		})
	}
}
