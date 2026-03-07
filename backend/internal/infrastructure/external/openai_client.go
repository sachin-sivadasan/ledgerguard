package external

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sachin-sivadasan/ledgerguard/internal/chat"
	openai "github.com/sashabaranov/go-openai"
)

// OpenAIClient implements chat.AIClient using the OpenAI API.
type OpenAIClient struct {
	client *openai.Client
	model  string
}

// NewOpenAIClient creates a new OpenAI chat client.
func NewOpenAIClient(apiKey, model string) *OpenAIClient {
	if model == "" {
		model = "gpt-4o"
	}
	return &OpenAIClient{
		client: openai.NewClient(apiKey),
		model:  model,
	}
}

func (c *OpenAIClient) ChatCompletion(ctx context.Context, req chat.ChatCompletionRequest) (*chat.ChatCompletionResponse, error) {
	messages := convertMessages(req.SystemPrompt, req.Messages)
	tools := convertTools(req.Tools)

	model := req.Model
	if model == "" {
		model = c.model
	}

	createReq := openai.ChatCompletionRequest{
		Model:    model,
		Messages: messages,
	}
	if len(tools) > 0 {
		createReq.Tools = tools
	}

	resp, err := c.client.CreateChatCompletion(ctx, createReq)
	if err != nil {
		return nil, fmt.Errorf("openai chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("openai: no choices in response")
	}

	choice := resp.Choices[0]
	result := &chat.ChatCompletionResponse{
		Content: choice.Message.Content,
		Usage: chat.TokenUsage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}

	for _, tc := range choice.Message.ToolCalls {
		result.ToolCalls = append(result.ToolCalls, chat.ToolCallRequest{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: tc.Function.Arguments,
		})
	}

	return result, nil
}

func convertMessages(systemPrompt string, messages []chat.ChatMessage) []openai.ChatCompletionMessage {
	var result []openai.ChatCompletionMessage

	if systemPrompt != "" {
		result = append(result, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		})
	}

	for _, m := range messages {
		msg := openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		}

		// Convert tool calls from assistant messages
		for _, tc := range m.ToolCalls {
			msg.ToolCalls = append(msg.ToolCalls, openai.ToolCall{
				ID:   tc.ID,
				Type: openai.ToolTypeFunction,
				Function: openai.FunctionCall{
					Name:      tc.Name,
					Arguments: tc.Arguments,
				},
			})
		}

		// Tool result messages
		if m.ToolResult != nil {
			msg.Role = openai.ChatMessageRoleTool
			msg.Content = m.ToolResult.Content
			msg.ToolCallID = m.ToolResult.ToolCallID
		}

		result = append(result, msg)
	}

	return result
}

func convertTools(tools []chat.ToolDefinition) []openai.Tool {
	result := make([]openai.Tool, 0, len(tools))
	for _, t := range tools {
		var params json.RawMessage
		if t.Parameters != nil {
			params = t.Parameters
		} else {
			params = json.RawMessage(`{"type":"object","properties":{}}`)
		}

		result = append(result, openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  params,
			},
		})
	}
	return result
}
