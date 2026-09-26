package mistral

import (
	"context"
	"errors"
	"fmt"

	"charm.land/fantasy"
	"github.com/charmbracelet/openai-go"
	"github.com/charmbracelet/openai-go/packages/param"
)

// mistralLanguageModel is the Mistral implementation of the LanguageModel interface.
// It reuses the OpenAI client since Mistral's API is OpenAI-compatible.
type mistralLanguageModel struct {
	modelID string
	client  openai.Client
}

// newMistralLanguageModel creates a new Mistral language model.
func newMistralLanguageModel(modelID string, client openai.Client) fantasy.LanguageModel {
	return &mistralLanguageModel{
		modelID: modelID,
		client:  client,
	}
}

// Generate implements fantasy.LanguageModel.
func (m *mistralLanguageModel) Generate(ctx context.Context, call fantasy.Call) (*fantasy.Response, error) {
	messages, _ := DefaultToPrompt(call.Prompt, Name, m.modelID)

	// Create params
	params := &openai.ChatCompletionNewParams{
		Messages: messages,
		Model:    m.modelID,
	}

	if call.MaxOutputTokens != nil {
		params.MaxTokens = param.NewOpt(*call.MaxOutputTokens)
	}
	if call.Temperature != nil {
		params.Temperature = param.NewOpt(*call.Temperature)
	}
	if call.TopP != nil {
		params.TopP = param.NewOpt(*call.TopP)
	}

	resp, err := m.client.Chat.Completions.New(ctx, *params)
	if err != nil {
		return nil, toMistralErr(err)
	}

	if resp == nil {
		return nil, errors.New("provider returned nil response")
	}

	if len(resp.Choices) == 0 {
		return nil, errors.New("no response generated")
	}

	choice := resp.Choices[0]
	content := make([]fantasy.Content, 0, 1)
	text := choice.Message.Content
	if text != "" {
		content = append(content, fantasy.TextContent{
			Text: text,
		})
	}

	// Handle tool calls
	for _, tc := range choice.Message.ToolCalls {
		content = append(content, fantasy.ToolCallContent{
			ProviderExecuted: false,
			ToolCallID:       tc.ID,
			ToolName:         tc.Function.Name,
			Input:            tc.Function.Arguments,
		})
	}

	return &fantasy.Response{
		Content:         content,
		Usage:           fantasy.Usage{TotalTokens: resp.Usage.TotalTokens},
		FinishReason:    mapFinishReason(choice.FinishReason),
		ProviderMetadata: fantasy.ProviderMetadata{},
	}, nil
}

// Stream implements fantasy.LanguageModel.
func (m *mistralLanguageModel) Stream(ctx context.Context, call fantasy.Call) (fantasy.StreamResponse, error) {
	messages, _ := DefaultToPrompt(call.Prompt, Name, m.modelID)

	// Create params
	params := &openai.ChatCompletionNewParams{
		Messages: messages,
		Model:    m.modelID,
	}

	if call.MaxOutputTokens != nil {
		params.MaxTokens = param.NewOpt(*call.MaxOutputTokens)
	}
	if call.Temperature != nil {
		params.Temperature = param.NewOpt(*call.Temperature)
	}
	if call.TopP != nil {
		params.TopP = param.NewOpt(*call.TopP)
	}

	params.StreamOptions = openai.ChatCompletionStreamOptionsParam{
		IncludeUsage: openai.Bool(true),
	}

	stream := m.client.Chat.Completions.NewStreaming(ctx, *params)
	return func(yield func(fantasy.StreamPart) bool) {
		for stream.Next() {
			chunk := stream.Current()
			if len(chunk.Choices) == 0 {
				continue
			}

			for _, choice := range chunk.Choices {
				if choice.Delta.Content != "" {
					if !yield(fantasy.StreamPart{
						Type:  fantasy.StreamPartTypeTextDelta,
						ID:    "0",
						Delta: choice.Delta.Content,
					}) {
						return
					}
				}

				if choice.FinishReason != "" {
					if !yield(fantasy.StreamPart{
						Type:         fantasy.StreamPartTypeFinish,
						FinishReason: mapFinishReason(choice.FinishReason),
					}) {
						return
					}
				}
			}
		}
	}, nil
}

// GenerateObject implements fantasy.LanguageModel.
func (m *mistralLanguageModel) GenerateObject(ctx context.Context, call fantasy.ObjectCall) (*fantasy.ObjectResponse, error) {
	// For Mistral, we reuse Generate and parse the response
	resp, err := m.Generate(ctx, fantasy.Call{
		Prompt:           call.Prompt,
		MaxOutputTokens:  call.MaxOutputTokens,
		Temperature:      call.Temperature,
		TopP:             call.TopP,
		TopK:             call.TopK,
		PresencePenalty:  call.PresencePenalty,
		FrequencyPenalty: call.FrequencyPenalty,
	})
	if err != nil {
		return nil, err
	}

	return &fantasy.ObjectResponse{
		Object:       resp.Content.Text(),
		RawText:      resp.Content.Text(),
		Usage:        resp.Usage,
		FinishReason: resp.FinishReason,
		Warnings:     resp.Warnings,
	}, nil
}

// StreamObject implements fantasy.LanguageModel.
func (m *mistralLanguageModel) StreamObject(ctx context.Context, call fantasy.ObjectCall) (fantasy.ObjectStreamResponse, error) {
	// For Mistral, we reuse Stream and convert to ObjectStreamResponse
	stream, err := m.Stream(ctx, fantasy.Call{
		Prompt:           call.Prompt,
		MaxOutputTokens:  call.MaxOutputTokens,
		Temperature:      call.Temperature,
		TopP:             call.TopP,
		TopK:             call.TopK,
		PresencePenalty:  call.PresencePenalty,
		FrequencyPenalty: call.FrequencyPenalty,
	})
	if err != nil {
		return nil, err
	}

	return func(yield func(fantasy.ObjectStreamPart) bool) {
		for part := range stream {
			if !yield(fantasy.ObjectStreamPart{
				Type:  fantasy.ObjectStreamPartTypeTextDelta,
				Delta: part.Delta,
			}) {
				return
			}
		}
	}, nil
}

// Provider implements fantasy.LanguageModel.
func (m *mistralLanguageModel) Provider() string {
	return Name
}

// Model implements fantasy.LanguageModel.
func (m *mistralLanguageModel) Model() string {
	return m.modelID
}

// mapFinishReason maps Mistral finish reasons to Fantasy finish reasons.
func mapFinishReason(reason string) fantasy.FinishReason {
	switch reason {
	case "stop":
		return "end_turn"
	case "length":
		return "length"
	case "tool_calls":
		return "tool_calls"
	default:
		return "unknown"
	}
}

// toMistralErr converts OpenAI errors to Mistral errors.
func toMistralErr(err error) error {
	var oaiErr *openai.Error
	if errors.As(err, &oaiErr) {
		return fmt.Errorf("Mistral API error: %s", oaiErr.Message)
	}

	return err
}