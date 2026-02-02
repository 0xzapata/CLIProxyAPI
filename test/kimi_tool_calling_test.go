package test

import (
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v6/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestKimiToolCallParser_ConcatenatedToolCalls(t *testing.T) {
	parser := util.NewKimiToolCallParser()

	// Test case from vLLM PR #24847 - concatenated without spacing
	input := `<|tool_calls_section_begin|><|tool_call_begin|>functions.get_weather:0<|tool_call_argument_begin|>{"city":"Paris"}<|tool_call_end|><|tool_call_begin|>functions.get_weather:1<|tool_call_argument_begin|>{"city":"London"}<|tool_call_end|><|tool_calls_section_end|>`

	calls := parser.ParseToolCalls(input)

	assert.Len(t, calls, 2, "Should parse both tool calls")

	// First call
	assert.Equal(t, "functions.get_weather:0", calls[0].ID)
	assert.Equal(t, "functions.get_weather", calls[0].Name)
	assert.Equal(t, `{"city":"Paris"}`, calls[0].Arguments)

	// Second call
	assert.Equal(t, "functions.get_weather:1", calls[1].ID)
	assert.Equal(t, "functions.get_weather", calls[1].Name)
	assert.Equal(t, `{"city":"London"}`, calls[1].Arguments)
}

func TestKimiToolCallParser_WithNewlinesInJSON(t *testing.T) {
	parser := util.NewKimiToolCallParser()

	// Test with pretty-printed JSON containing newlines
	input := `<|tool_calls_section_begin|><|tool_call_begin|>functions.test:0<|tool_call_argument_begin|>{
  "name": "test",
  "value": 123
}<|tool_call_end|><|tool_calls_section_end|>`

	calls := parser.ParseToolCalls(input)

	assert.Len(t, calls, 1)
	assert.Equal(t, "functions.test:0", calls[0].ID)
	assert.Equal(t, "functions.test", calls[0].Name)
	assert.Contains(t, calls[0].Arguments, `"name": "test"`)
	assert.Contains(t, calls[0].Arguments, `"value": 123`)
}

func TestKimiToolCallParser_SingleToolCall(t *testing.T) {
	parser := util.NewKimiToolCallParser()

	input := `<|tool_calls_section_begin|><|tool_call_begin|>functions.search:0<|tool_call_argument_begin|>{"query":"test"}<|tool_call_end|><|tool_calls_section_end|>`

	calls := parser.ParseToolCalls(input)

	assert.Len(t, calls, 1)
	assert.Equal(t, "functions.search:0", calls[0].ID)
	assert.Equal(t, "functions.search", calls[0].Name)
	assert.Equal(t, `{"query":"test"}`, calls[0].Arguments)
}

func TestKimiToolCallParser_NoToolCalls(t *testing.T) {
	parser := util.NewKimiToolCallParser()

	input := `This is just regular text with no tool calls.`

	calls := parser.ParseToolCalls(input)

	assert.Len(t, calls, 0)
}

func TestKimiToolCallParser_EmptyInput(t *testing.T) {
	parser := util.NewKimiToolCallParser()

	calls := parser.ParseToolCalls("")

	assert.Len(t, calls, 0)
}

func TestKimiToolCallParser_WithSpacesAroundTokens(t *testing.T) {
	parser := util.NewKimiToolCallParser()

	// Test with extra spaces around tokens (should still parse)
	input := `<|tool_calls_section_begin|> <|tool_call_begin|> functions.get_weather:0 <|tool_call_argument_begin|> {"city":"Paris"} <|tool_call_end|> <|tool_calls_section_end|>`

	calls := parser.ParseToolCalls(input)

	assert.Len(t, calls, 1)
	assert.Equal(t, "functions.get_weather:0", calls[0].ID)
	assert.Equal(t, "functions.get_weather", calls[0].Name)
	assert.Equal(t, `{"city":"Paris"}`, calls[0].Arguments)
}

func TestKimiToolCallParser_ComplexJSONArguments(t *testing.T) {
	parser := util.NewKimiToolCallParser()

	input := `<|tool_calls_section_begin|><|tool_call_begin|>functions.complex:0<|tool_call_argument_begin|>{"nested":{"key":"value"},"array":[1,2,3],"string":"test"}<|tool_call_end|><|tool_calls_section_end|>`

	calls := parser.ParseToolCalls(input)

	assert.Len(t, calls, 1)
	assert.Equal(t, "functions.complex:0", calls[0].ID)
	assert.Equal(t, "functions.complex", calls[0].Name)
	assert.Equal(t, `{"nested":{"key":"value"},"array":[1,2,3],"string":"test"}`, calls[0].Arguments)
}

func TestKimiToolCallParser_MultipleFunctions(t *testing.T) {
	parser := util.NewKimiToolCallParser()

	input := `<|tool_calls_section_begin|><|tool_call_begin|>functions.search:0<|tool_call_argument_begin|>{"q":"test"}<|tool_call_end|><|tool_call_begin|>functions.calculate:1<|tool_call_argument_begin|>{"expr":"2+2"}<|tool_call_end|><|tool_call_begin|>functions.format:2<|tool_call_argument_begin|>{"text":"result"}<|tool_call_end|><|tool_calls_section_end|>`

	calls := parser.ParseToolCalls(input)

	assert.Len(t, calls, 3)

	// Verify order and content
	assert.Equal(t, "functions.search:0", calls[0].ID)
	assert.Equal(t, "functions.search", calls[0].Name)
	assert.Equal(t, `{"q":"test"}`, calls[0].Arguments)

	assert.Equal(t, "functions.calculate:1", calls[1].ID)
	assert.Equal(t, "functions.calculate", calls[1].Name)
	assert.Equal(t, `{"expr":"2+2"}`, calls[1].Arguments)

	assert.Equal(t, "functions.format:2", calls[2].ID)
	assert.Equal(t, "functions.format", calls[2].Name)
	assert.Equal(t, `{"text":"result"}`, calls[2].Arguments)
}

func TestKimiToolCallParser_WithSpecialCharactersInJSON(t *testing.T) {
	parser := util.NewKimiToolCallParser()

	input := `<|tool_calls_section_begin|><|tool_call_begin|>functions.escape:0<|tool_call_argument_begin|>{"text":"Hello \"World\" \n\t"}<|tool_call_end|><|tool_calls_section_end|>`

	calls := parser.ParseToolCalls(input)

	assert.Len(t, calls, 1)
	assert.Equal(t, "functions.escape:0", calls[0].ID)
	assert.Equal(t, "functions.escape", calls[0].Name)
	assert.Equal(t, `{"text":"Hello \"World\" \n\t"}`, calls[0].Arguments)
}

func TestKimiToolCallParser_IndexFormat(t *testing.T) {
	parser := util.NewKimiToolCallParser()

	// Test that the index format (name:index) is correctly parsed
	input := `<|tool_calls_section_begin|><|tool_call_begin|>my_function:42<|tool_call_argument_begin|>{"arg":"value"}<|tool_call_end|><|tool_calls_section_end|>`

	calls := parser.ParseToolCalls(input)

	assert.Len(t, calls, 1)
	assert.Equal(t, "my_function:42", calls[0].ID)
	assert.Equal(t, "my_function", calls[0].Name)
}
