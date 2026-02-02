// Package util provides utility functions for processing non-standard reasoning fields
// in API responses, such as Kimi K2.5's "reasoning_content" field.
package util

import (
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// InterleavedConfig holds the configuration for processing non-standard reasoning fields.
type InterleavedConfig struct {
	Field           string
	FormatAsContent bool
	Separator       string
}

// ProcessInterleavedReasoning processes responses with non-standard reasoning fields
// like Kimi K2.5's "reasoning_content".
func ProcessInterleavedReasoning(body []byte, config *InterleavedConfig) []byte {
	if config == nil || config.Field == "" {
		return body
	}

	// Check if the reasoning field exists
	reasoningPath := "choices.0.message." + config.Field
	reasoning := gjson.GetBytes(body, reasoningPath)
	if !reasoning.Exists() || reasoning.String() == "" {
		return body
	}

	contentPath := "choices.0.message.content"
	content := gjson.GetBytes(body, contentPath)

	// Build the new content
	var newContent strings.Builder
	newContent.WriteString("Thinking...\n\n")
	newContent.WriteString(reasoning.String())

	if config.FormatAsContent {
		if config.Separator != "" {
			newContent.WriteString(config.Separator)
		} else {
			newContent.WriteString("\n\n")
		}
	}

	if content.Exists() && content.String() != "" {
		newContent.WriteString(content.String())
	}

	// Update the response
	result, _ := sjson.SetBytes(body, contentPath, newContent.String())

	// Remove the reasoning field from output (we've merged it into content)
	result, _ = sjson.DeleteBytes(result, reasoningPath)

	return result
}

// ProcessInterleavedReasoningStream handles streaming SSE chunks with reasoning_content.
func ProcessInterleavedReasoningStream(chunk []byte, config *InterleavedConfig) []byte {
	if config == nil || config.Field == "" {
		return chunk
	}

	// Parse SSE line
	line := strings.TrimSpace(string(chunk))
	if !strings.HasPrefix(line, "data:") {
		return chunk
	}

	jsonStr := strings.TrimPrefix(line, "data:")
	jsonStr = strings.TrimSpace(jsonStr)
	if jsonStr == "[DONE]" {
		return chunk
	}

	// Check for reasoning_content in delta
	reasoningPath := "choices.0.delta." + config.Field
	reasoning := gjson.Get(jsonStr, reasoningPath)

	if reasoning.Exists() && reasoning.String() != "" {
		// Format reasoning content with proper spacing
		separator := config.Separator
		if separator == "" {
			separator = "\n\n"
		}
		formatted := "Thinking...\n\n" + reasoning.String() + separator

		// Create new delta with content
		result, _ := sjson.Set(jsonStr, "choices.0.delta.content", formatted)
		result, _ = sjson.Delete(result, reasoningPath)

		return []byte("data: " + result + "\n\n")
	}

	// Preserve other deltas (tool_calls, content, etc.)
	return chunk
}
