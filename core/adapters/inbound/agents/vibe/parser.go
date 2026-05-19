package vibe

import (
	"strings"

	"irrlicht/core/pkg/tailer"
)

// Parser implements tailer.TranscriptParser for Mistral Vibe transcripts.
//
// Vibe uses the Agent Client Protocol (ACP) and stores messages in JSONL format
// with each line representing a serialized LLMMessage. The message structure includes:
//   - role: "system", "user", "assistant", or "tool"
//   - content: message text content
//   - tool_calls: array of tool call objects (for assistant messages)
//   - tool_call_id: identifier for tool result messages
//   - name: tool name (for tool messages)
//   - reasoning_content: reasoning/think content
//
// Vibe session directories contain a messages.jsonl file with append-only
// message logs. Each message is a complete JSON object on a single line.
type Parser struct{}

// ParseLine parses a Mistral Vibe JSONL line into a normalized ParsedEvent.
func (p *Parser) ParseLine(raw map[string]interface{}) *tailer.ParsedEvent {
	ev := &tailer.ParsedEvent{
		Timestamp: tailer.ParseTimestamp(raw),
	}

	// Extract role from the message
	role, _ := raw["role"].(string)

	// Handle metadata fields that aren't message content
	if model, ok := raw["model"].(string); ok && model != "" {
		ev.ModelName = tailer.NormalizeModelName(model)
	}

	// Extract usage information if present (Vibe may include usage in some events)
	if usage, ok := raw["usage"].(map[string]interface{}); ok {
		ev.Tokens = extractVibeUsage(usage)
	}

	// Classify and process based on role
	switch role {
	case "system":
		// System messages are typically the initial prompt - skip them
		// but record model info if present
		ev.Skip = true
		if ev.ModelName == "" {
			// Try to extract model from content or other fields
			if content, ok := raw["content"].(string); ok {
				if model := extractModelFromContent(content); model != "" {
					ev.ModelName = model
				}
			}
		}
		return ev

	case "user":
		ev.EventType = "user"
		ev.ClearToolNames = true
		// Extract assistant text from content for display
		if content, ok := raw["content"].(string); ok {
			if strings.Contains(content, "[Request interrupted") {
				ev.IsUserInterrupt = true
			}
			ev.ContentChars = int64(len(content))
		}

	case "assistant":
		// Check for tool calls
		if toolCalls, ok := raw["tool_calls"].([]interface{}); ok && len(toolCalls) > 0 {
			ev.EventType = "assistant"
			for _, tcInterface := range toolCalls {
				if tc, ok := tcInterface.(map[string]interface{}); ok {
					if name, ok := tc["function"].(map[string]interface{})["name"].(string); ok && name != "" {
						if id, ok := tc["id"].(string); ok && id != "" {
							ev.ToolUses = append(ev.ToolUses, tailer.ToolUse{ID: id, Name: name})
						} else {
							ev.ToolUses = append(ev.ToolUses, tailer.ToolUse{ID: "", Name: name})
						}
					}
				}
			}
		} else {
			// Regular assistant message without tool calls
			ev.EventType = "assistant"
		}
		
		// Extract reasoning content if present
		if reasoning, ok := raw["reasoning_content"].(string); ok && reasoning != "" {
			// Reasoning content is part of the assistant's work
			if ev.AssistantText == "" {
				ev.AssistantText = reasoning
			}
		}
		
		// Extract regular content
		if content, ok := raw["content"].(string); ok && content != "" {
			if ev.AssistantText == "" {
				ev.AssistantText = content
			} else {
				ev.AssistantText += "\n" + content
			}
			ev.ContentChars = int64(len(content))
		}
		
		// Check for stop_reason or other end-of-turn indicators
		// Vibe doesn't always include stop_reason, so we check for content completion
		if stopped, ok := raw["stopped_by_middleware"].(bool); ok && stopped {
			// Middleware stopped the response - might be end of turn
			// but not necessarily final
		}
		
		// For now, we don't have a clear turn_done indicator from Vibe
		// The fswatcher will detect file closes as activity end

	case "tool":
		// Tool result message
		ev.EventType = "tool_result"
		if toolCallID, ok := raw["tool_call_id"].(string); ok && toolCallID != "" {
			ev.ToolResultIDs = []string{toolCallID}
		}
		if name, ok := raw["name"].(string); ok && name != "" {
			// Record the tool name for tracking
			if len(ev.ToolUses) == 0 {
				ev.ToolUses = append(ev.ToolUses, tailer.ToolUse{ID: "", Name: name})
			}
		}
		if content, ok := raw["content"].(string); ok {
			ev.ContentChars = int64(len(content))
		}
		if isErr, ok := raw["is_error"].(bool); ok && isErr {
			ev.IsError = true
		}

	default:
		// Unknown role - skip but log content chars
		ev.Skip = true
		if content, ok := raw["content"].(string); ok {
			ev.ContentChars = int64(len(content))
		}
	}

	// Truncate assistant text for display
	if len([]rune(ev.AssistantText)) > 200 {
		runes := []rune(ev.AssistantText)
		ev.AssistantText = string(runes[:200])
	}

	return ev
}

// extractVibeUsage extracts token usage from a Vibe usage object.
// Vibe usage typically has prompt_tokens and completion_tokens.
func extractVibeUsage(usage map[string]interface{}) *tailer.TokenSnapshot {
	snapshot := &tailer.TokenSnapshot{}

	if v, ok := usage["prompt_tokens"].(float64); ok {
		snapshot.Input = int64(v)
	}
	if v, ok := usage["completion_tokens"].(float64); ok {
		snapshot.Output = int64(v)
	}
	if v, ok := usage["total_tokens"].(float64); ok {
		// If total is provided but not input/output, estimate
		if snapshot.Input == 0 && snapshot.Output == 0 {
			snapshot.Total = int64(v)
		}
	}

	// Return nil if no tokens were found
	if snapshot.Input == 0 && snapshot.Output == 0 && snapshot.Total == 0 {
		return nil
	}
	return snapshot
}

// extractModelFromContent attempts to extract model name from system content.
// Vibe system prompts often include model information.
func extractModelFromContent(content string) string {
	// Look for common model name patterns in the content
	// Vibe doesn't typically include model in the message itself,
	// but we check anyway for robustness
	lower := strings.ToLower(content)
	
	// Common Mistral model names
	models := []string{
		"mistral-large",
		"mistral-medium",
		"mistral-small",
		"devstral",
		"codestral",
		"mistral-vibe",
	}
	
	for _, model := range models {
		if strings.Contains(lower, model) {
			return model
		}
	}
	
	return ""
}
