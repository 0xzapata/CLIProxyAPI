// Package util provides parsing utilities for Kimi K2.5's special tool call format.
// Tool calls can be concatenated without proper spacing, requiring special handling.
package util

import (
	"regexp"
	"strings"
)

// KimiToolCall represents a parsed tool call from Kimi K2.5 format.
type KimiToolCall struct {
	ID        string
	Name      string
	Arguments string
}

// KimiToolCallParser handles Kimi K2.5's special tool call format.
// Tool calls can be concatenated without proper spacing:
// <|tool_calls_section_begin|><|tool_call_begin|>func:0<|tool_call_argument_begin|>{"arg":"val"}<|tool_call_end|>
//
// Based on vLLM PR #24847: https://github.com/vllm-project/vllm/pull/24847
type KimiToolCallParser struct {
	toolCallRegex *regexp.Regexp
}

// NewKimiToolCallParser creates a new parser for Kimi K2.5 tool calls.
//
// Note: Go's RE2 doesn't support negative lookahead, so we use a different approach:
// We match everything between markers and then validate that arguments don't contain
// the next tool call marker.
func NewKimiToolCallParser() *KimiToolCallParser {
	// Pattern for Kimi K2.5 tool calls
	// Matches: <|tool_call_begin|>ID:INDEX<|tool_call_argument_begin|>ARGUMENTS<|tool_call_end|>
	// We use [\s\S]*? to match any character (including newlines) non-greedily
	pattern := regexp.MustCompile(
		`<\|tool_call_begin\|>\s*([^<:\s][^<:\s]*?:\d+)\s*<\|tool_call_argument_begin\|>\s*([\s\S]*?)\s*<\|tool_call_end\|>`,
	)
	return &KimiToolCallParser{
		toolCallRegex: pattern,
	}
}

// ParseToolCalls extracts tool calls from Kimi K2.5 content.
func (p *KimiToolCallParser) ParseToolCalls(content string) []KimiToolCall {
	matches := p.toolCallRegex.FindAllStringSubmatch(content, -1)

	calls := make([]KimiToolCall, 0, len(matches))
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		id := strings.TrimSpace(match[1])
		args := match[2]

		// Validate that arguments don't contain another tool_call_begin marker
		// This prevents matching across tool call boundaries
		if strings.Contains(args, "<|tool_call_begin|>") {
			continue
		}

		// Extract function name from ID (format: "function_name:call_index")
		parts := strings.Split(id, ":")
		if len(parts) >= 2 {
			funcName := strings.TrimSpace(parts[0])
			calls = append(calls, KimiToolCall{
				ID:        id,
				Name:      funcName,
				Arguments: args,
			})
		}
	}

	return calls
}
