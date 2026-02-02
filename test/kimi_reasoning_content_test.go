package test

import (
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v6/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestProcessInterleavedReasoning(t *testing.T) {
	config := &util.InterleavedConfig{
		Field:           "reasoning_content",
		FormatAsContent: true,
		Separator:       "\n\n",
	}

	input := []byte(`{
		"choices": [{
			"message": {
				"role": "assistant",
				"reasoning_content": "The user is saying hello.",
				"content": "Hello! How can I help you today?"
			}
		}]
	}`)

	result := util.ProcessInterleavedReasoning(input, config)

	// Verify reasoning_content is merged into content with proper formatting
	resultStr := string(result)
	assert.Contains(t, resultStr, "Thinking...")
	assert.Contains(t, resultStr, "The user is saying hello.")
	assert.Contains(t, resultStr, "Hello! How can I help you today?")
	assert.NotContains(t, resultStr, "reasoning_content")
}

func TestProcessInterleavedReasoning_NoReasoningField(t *testing.T) {
	config := &util.InterleavedConfig{
		Field:           "reasoning_content",
		FormatAsContent: true,
		Separator:       "\n\n",
	}

	input := []byte(`{
		"choices": [{
			"message": {
				"role": "assistant",
				"content": "Hello! How can I help you today?"
			}
		}]
	}`)

	result := util.ProcessInterleavedReasoning(input, config)

	// Should return unchanged when reasoning_content field doesn't exist
	resultStr := string(result)
	assert.Contains(t, resultStr, "Hello! How can I help you today?")
	assert.NotContains(t, resultStr, "Thinking...")
}

func TestProcessInterleavedReasoning_OnlyReasoningNoContent(t *testing.T) {
	config := &util.InterleavedConfig{
		Field:           "reasoning_content",
		FormatAsContent: true,
		Separator:       "\n\n",
	}

	input := []byte(`{
		"choices": [{
			"message": {
				"role": "assistant",
				"reasoning_content": "Let me think about this..."
			}
		}]
	}`)

	result := util.ProcessInterleavedReasoning(input, config)

	// Should format reasoning with Thinking... prefix
	resultStr := string(result)
	assert.Contains(t, resultStr, "Thinking...")
	assert.Contains(t, resultStr, "Let me think about this...")
	assert.NotContains(t, resultStr, "reasoning_content")
}

func TestProcessInterleavedReasoningStream(t *testing.T) {
	config := &util.InterleavedConfig{
		Field:           "reasoning_content",
		FormatAsContent: true,
		Separator:       "\n\n",
	}

	// Test reasoning_content delta chunk
	input := []byte(`data: {"choices":[{"delta":{"reasoning_content":"Let me think about this..."}}]}`)

	result := util.ProcessInterleavedReasoningStream(input, config)

	resultStr := string(result)
	assert.Contains(t, resultStr, "Thinking...")
	assert.Contains(t, resultStr, "Let me think about this...")
	assert.NotContains(t, resultStr, "reasoning_content")
}

func TestProcessInterleavedReasoningStream_Done(t *testing.T) {
	config := &util.InterleavedConfig{
		Field:           "reasoning_content",
		FormatAsContent: true,
		Separator:       "\n\n",
	}

	input := []byte(`data: [DONE]`)

	result := util.ProcessInterleavedReasoningStream(input, config)

	// [DONE] chunks should pass through unchanged
	assert.Equal(t, string(input), string(result))
}

func TestProcessInterleavedReasoningStream_NonDataLine(t *testing.T) {
	config := &util.InterleavedConfig{
		Field:           "reasoning_content",
		FormatAsContent: true,
		Separator:       "\n\n",
	}

	input := []byte(`: this is a comment line`)

	result := util.ProcessInterleavedReasoningStream(input, config)

	// Non-data lines should pass through unchanged
	assert.Equal(t, string(input), string(result))
}

func TestProcessInterleavedReasoningStream_RegularContent(t *testing.T) {
	config := &util.InterleavedConfig{
		Field:           "reasoning_content",
		FormatAsContent: true,
		Separator:       "\n\n",
	}

	// Test regular content delta without reasoning_content
	input := []byte(`data: {"choices":[{"delta":{"content":"Hello world"}}]}`)

	result := util.ProcessInterleavedReasoningStream(input, config)

	// Regular content should pass through unchanged
	assert.Equal(t, string(input), string(result))
}

func TestProcessInterleavedReasoning_NoConfig(t *testing.T) {
	// No config should return body unchanged
	input := []byte(`{
		"choices": [{
			"message": {
				"reasoning_content": "The user is saying hello.",
				"content": "Hello!"
			}
		}]
	}`)

	result := util.ProcessInterleavedReasoning(input, nil)

	assert.Equal(t, string(input), string(result))
}

func TestProcessInterleavedReasoning_EmptyField(t *testing.T) {
	// Empty field name should return body unchanged
	config := &util.InterleavedConfig{
		Field:           "",
		FormatAsContent: true,
		Separator:       "\n\n",
	}

	input := []byte(`{
		"choices": [{
			"message": {
				"reasoning_content": "The user is saying hello.",
				"content": "Hello!"
			}
		}]
	}`)

	result := util.ProcessInterleavedReasoning(input, config)

	assert.Equal(t, string(input), string(result))
}

func TestProcessInterleavedReasoning_NoFormatAsContent(t *testing.T) {
	config := &util.InterleavedConfig{
		Field:           "reasoning_content",
		FormatAsContent: false,
		Separator:       "\n\n",
	}

	input := []byte(`{
		"choices": [{
			"message": {
				"reasoning_content": "The user is saying hello.",
				"content": "Hello!"
			}
		}]
	}`)

	result := util.ProcessInterleavedReasoning(input, config)

	// When FormatAsContent is false, reasoning should still be merged
	resultStr := string(result)
	assert.Contains(t, resultStr, "Thinking...")
	assert.Contains(t, resultStr, "The user is saying hello.")
	assert.NotContains(t, resultStr, "reasoning_content")
}

func TestProcessInterleavedReasoning_EmptySeparator(t *testing.T) {
	config := &util.InterleavedConfig{
		Field:           "reasoning_content",
		FormatAsContent: true,
		Separator:       "",
	}

	input := []byte(`{
		"choices": [{
			"message": {
				"reasoning_content": "Thinking...",
				"content": "Response"
			}
		}]
	}`)

	result := util.ProcessInterleavedReasoning(input, config)

	// Empty separator should default to "\n\n"
	// Note: JSON output has escaped newlines (\n becomes \\n in JSON string)
	resultStr := string(result)
	assert.Contains(t, resultStr, `Thinking...\n\n`)
	assert.Contains(t, resultStr, "Response")
}
