package mistral

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"charm.land/fantasy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMistralGenerate tests the Generate method of the Mistral language model.
func TestMistralGenerate(t *testing.T) {
	t.Parallel()

	t.Run("should return a response for a simple prompt", func(t *testing.T) {
		t.Parallel()

		// Mock server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/v1/chat/completions", r.URL.Path)
			assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

			// Validate request body
			var body map[string]any
			reqBody, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			require.NoError(t, json.Unmarshal(reqBody, &body))

			assert.Equal(t, "mistral-tiny", body["model"])
			assert.Len(t, body["messages"].([]any), 1)

			// Return a mock response
			w.WriteHeader(http.StatusOK)
			resp := map[string]any{
				"id":      "chatcmpl-test-id",
				"object":  "chat.completion",
				"created": 1234567890,
				"model":   "mistral-tiny",
				"choices": []map[string]any{
					{
						"index": 0,
						"message": map[string]any{
							"role":    "assistant",
							"content": "Hello! How can I help you today?",
						},
						"finish_reason": "stop",
					},
				},
				"usage": map[string]any{
					"prompt_tokens":     int64(10),
					"completion_tokens": int64(15),
					"total_tokens":      int64(25),
				},
			}
			require.NoError(t, json.NewEncoder(w).Encode(resp))
		}))
		defer server.Close()

		// Create a Mistral provider
		provider, err := NewProvider(WithAPIKey("test-api-key"), WithBaseURL(server.URL))
		require.NoError(t, err)

		// Create a language model
		lm := provider.LanguageModel("mistral-tiny")

		// Create a call
		call := fantasy.Call{
			Prompt: fantasy.Prompt{
				{
					Role: fantasy.MessageRoleUser,
					Content: []fantasy.MessagePart{
						fantasy.TextPart{Text: "Hello"},
					},
				},
			},
		}

		// Call Generate
		resp, err := lm.Generate(context.Background(), call)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Validate response
		require.Len(t, resp.Content, 1)
		assert.Equal(t, "Hello! How can I help you today?", resp.Content[0].(fantasy.TextContent).Text)
		assert.Equal(t, int64(25), resp.Usage.TotalTokens)
	})

	t.Run("should handle errors from the Mistral API", func(t *testing.T) {
		t.Parallel()

		// Mock server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": {"message": "invalid request", "type": "bad_request"}}`))
		}))
		defer server.Close()

		// Create a Mistral provider
		provider, err := NewProvider(WithAPIKey("test-api-key"), WithBaseURL(server.URL))
		require.NoError(t, err)

		// Create a language model
		lm := provider.LanguageModel("mistral-tiny")

		// Create a call
		call := fantasy.Call{
			Prompt: fantasy.Prompt{
				{
					Role: fantasy.MessageRoleUser,
					Content: []fantasy.MessagePart{
						fantasy.TextPart{Text: "Hello"},
					},
				},
			},
		}

		// Call Generate
		resp, err := lm.Generate(context.Background(), call)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

// TestMistralStream tests the Stream method of the Mistral language model.
func TestMistralStream(t *testing.T) {
	t.Parallel()

	t.Run("should stream a response for a simple prompt", func(t *testing.T) {
		t.Parallel()

		// Mock server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/v1/chat/completions", r.URL.Path)
			assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

			// Validate request body
			var body map[string]any
			reqBody, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			require.NoError(t, json.Unmarshal(reqBody, &body))

			assert.Equal(t, "mistral-tiny", body["model"])
			assert.Equal(t, true, body["stream"])

			// Return a streaming response
			w.WriteHeader(http.StatusOK)
			flusher, _ := w.(http.Flusher)

			// First chunk
			w.Write([]byte("data: {" +
				`"id":"chatcmpl-test-id","object":"chat.completion.chunk","created":1234567890,` +
				`"model":"mistral-tiny","choices":[{"index":0,"delta":{"role":"assistant"},"finish_reason":null}]` +
				"}\n\n"))
			flusher.Flush()

			// Second chunk
			w.Write([]byte("data: {" +
				`"id":"chatcmpl-test-id","object":"chat.completion.chunk","created":1234567890,` +
				`"model":"mistral-tiny","choices":[{"index":0,"delta":{"content":"Hello"},"finish_reason":null}]` +
				"}\n\n"))
			flusher.Flush()

			// Third chunk
			w.Write([]byte("data: {" +
				`"id":"chatcmpl-test-id","object":"chat.completion.chunk","created":1234567890,` +
				`"model":"mistral-tiny","choices":[{"index":0,"delta":{"content":"!"},"finish_reason":null}]` +
				"}\n\n"))
			flusher.Flush()

			// Fourth chunk (finish)
			w.Write([]byte("data: {" +
				`"id":"chatcmpl-test-id","object":"chat.completion.chunk","created":1234567890,` +
				`"model":"mistral-tiny","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]` +
				"}\n\n"))
			flusher.Flush()

			// Usage chunk
			w.Write([]byte("data: {" +
				`"id":"chatcmpl-test-id","object":"chat.completion.chunk","created":1234567890,` +
				`"model":"mistral-tiny","choices":[],"usage":{"prompt_tokens":10,"completion_tokens":15,"total_tokens":25}` +
				"}\n\ndata: [DONE]\n\n"))
			flusher.Flush()
		}))
		defer server.Close()

		// Create a Mistral provider
		provider, err := NewProvider(WithAPIKey("test-api-key"), WithBaseURL(server.URL))
		require.NoError(t, err)

		// Create a language model
		lm := provider.LanguageModel("mistral-tiny")

		// Create a call
		call := fantasy.Call{
			Prompt: fantasy.Prompt{
				{
					Role: fantasy.MessageRoleUser,
					Content: []fantasy.MessagePart{
						fantasy.TextPart{Text: "Hello"},
					},
				},
			},
		}

		// Call Stream
		stream := lm.Stream(context.Background(), call)
		require.NotNil(t, stream)

		// Collect stream parts
		var textParts []fantasy.StreamPart
		for part := range stream {
			textParts = append(textParts, part)
		}

		// Validate stream parts
		require.Len(t, textParts, 5) // 3 text deltas + 1 text end + 1 usage
		assert.Equal(t, fantasy.StreamPartTypeTextDelta, textParts[0].Type)
		assert.Equal(t, "Hello", textParts[0].Delta)
		assert.Equal(t, fantasy.StreamPartTypeTextDelta, textParts[1].Type)
		assert.Equal(t, "!", textParts[1].Delta)
		assert.Equal(t, fantasy.StreamPartTypeTextEnd, textParts[2].Type)
	})

	t.Run("should handle streaming errors from the Mistral API", func(t *testing.T) {
		t.Parallel()

		// Mock server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": {"message": "invalid request", "type": "bad_request"}}`))
		}))
		defer server.Close()

		// Create a Mistral provider
		provider, err := NewProvider(WithAPIKey("test-api-key"), WithBaseURL(server.URL))
		require.NoError(t, err)

		// Create a language model
		lm := provider.LanguageModel("mistral-tiny")

		// Create a call
		call := fantasy.Call{
			Prompt: fantasy.Prompt{
				{
					Role: fantasy.MessageRoleUser,
					Content: []fantasy.MessagePart{
						fantasy.TextPart{Text: "Hello"},
					},
				},
			},
		}

		// Call Stream
		stream, err := lm.Stream(context.Background(), call)
		assert.Error(t, err)
		assert.Nil(t, stream)
	})
}

// TestMistralLanguageModelProviderAndModel tests the Provider and Model methods.
func TestMistralLanguageModelProviderAndModel(t *testing.T) {
	t.Parallel()

	provider, err := NewProvider(WithAPIKey("test-api-key"))
	require.NoError(t, err)

	lm := provider.LanguageModel("mistral-tiny")
	assert.Equal(t, "mistral", lm.Provider())
	assert.Equal(t, "mistral-tiny", lm.Model())
}

// TestMistralGenerateObject tests the GenerateObject method.
func TestMistralGenerateObject(t *testing.T) {
	t.Parallel()

	t.Run("should return a structured response", func(t *testing.T) {
		t.Parallel()

		// Mock server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/v1/chat/completions", r.URL.Path)

			// Validate request body
			var body map[string]any
			reqBody, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			require.NoError(t, json.Unmarshal(reqBody, &body))

			assert.Equal(t, "mistral-tiny", body["model"])

			// Return a mock response
			w.WriteHeader(http.StatusOK)
			resp := map[string]any{
				"id":      "chatcmpl-test-id",
				"object":  "chat.completion",
				"created": int64(1234567890),
				"model":   "mistral-tiny",
				"choices": []map[string]any{
					{
						"index": 0,
						"message": map[string]any{
							"role":    "assistant",
							"content": `{"key": "value"}`,
						},
						"finish_reason": "stop",
					},
				},
				"usage": map[string]any{
					"prompt_tokens":     int64(10),
					"completion_tokens": int64(15),
					"total_tokens":      int64(25),
				},
			}
			require.NoError(t, json.NewEncoder(w).Encode(resp))
		}))
		defer server.Close()

		// Create a Mistral provider
		provider, err := NewProvider(WithAPIKey("test-api-key"), WithBaseURL(server.URL))
		require.NoError(t, err)

		// Create a language model
		lm := provider.LanguageModel("mistral-tiny")

		// Create an object call
		objCall := fantasy.ObjectCall{
			Call: fantasy.Call{
				Prompt: fantasy.Prompt{
					{
						Role: fantasy.MessageRoleUser,
						Content: []fantasy.MessagePart{
							fantasy.TextPart{Text: "Return a JSON object with a key 'key' and value 'value'"},
						},
					},
				},
			},
			Schema: &fantasy.Schema{
				Type: "object",
				Properties: map[string]*fantasy.Schema{
					"key": {Type: "string"},
				},
			},
		}

		// Call GenerateObject
		resp, err := lm.GenerateObject(context.Background(), objCall)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Validate response
		assert.Equal(t, `{"key":"value"}`, resp.Content.Text())
	})
}

// TestMistralStreamObject tests the StreamObject method.
func TestMistralStreamObject(t *testing.T) {
	t.Parallel()

	t.Run("should stream a structured response", func(t *testing.T) {
		t.Parallel()

		// Mock server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/v1/chat/completions", r.URL.Path)

			// Validate request body
			var body map[string]any
			reqBody, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			require.NoError(t, json.Unmarshal(reqBody, &body))

			assert.Equal(t, "mistral-tiny", body["model"])
			assert.Equal(t, true, body["stream"])

			// Return a streaming response
			w.WriteHeader(http.StatusOK)
			flusher, _ := w.(http.Flusher)

			// First chunk
			w.Write([]byte("data: {" +
				`"id":"chatcmpl-test-id","object":"chat.completion.chunk","created":1234567890,` +
				`"model":"mistral-tiny","choices":[{"index":0,"delta":{"role":"assistant"},"finish_reason":null}]` +
				"}\n\n"))
			flusher.Flush()

			// Second chunk (structured output)
			w.Write([]byte("data: {" +
				`"id":"chatcmpl-test-id","object":"chat.completion.chunk","created":1234567890,` +
				`"model":"mistral-tiny","choices":[{"index":0,"delta":{"content":"{\\"key\\": \\"value\\"}"},"finish_reason":null}]` +
				"}\n\n"))
			flusher.Flush()

			// Third chunk (finish)
			w.Write([]byte("data: {" +
				`"id":"chatcmpl-test-id","object":"chat.completion.chunk","created":1234567890,` +
				`"model":"mistral-tiny","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]` +
				"}\n\ndata: [DONE]\n\n"))
			flusher.Flush()
		}))
		defer server.Close()

		// Create a Mistral provider
		provider, err := NewProvider(WithAPIKey("test-api-key"), WithBaseURL(server.URL))
		require.NoError(t, err)

		// Create a language model
		lm := provider.LanguageModel("mistral-tiny")

		// Create an object call
		objCall := fantasy.ObjectCall{
			Call: fantasy.Call{
				Prompt: fantasy.Prompt{
					{
						Role: fantasy.MessageRoleUser,
						Content: []fantasy.MessagePart{
							fantasy.TextPart{Text: "Return a JSON object with a key 'key' and value 'value'"},
						},
					},
				},
			},
			Schema: &fantasy.Schema{
				Type: "object",
				Properties: map[string]*fantasy.Schema{
					"key": {Type: "string"},
				},
			},
		}

		// Call StreamObject
		stream := lm.StreamObject(context.Background(), objCall)
		require.NotNil(t, stream)

		// Collect stream parts
		var textParts []fantasy.StreamPart
		for part := range stream {
			textParts = append(textParts, part)
		}

		// Validate stream parts
		require.Len(t, textParts, 2) // 1 text delta + 1 finish
		assert.Equal(t, fantasy.StreamPartTypeTextDelta, textParts[0].Type)
		assert.Equal(t, `{"key": "value"}`, textParts[0].Delta)
	})
}

// TestMistralListModels tests listing available models.
func TestMistralListModels(t *testing.T) {
	t.Parallel()

	t.Run("should return a list of available models", func(t *testing.T) {
		t.Parallel()

		// Mock server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/v1/models", r.URL.Path)
			assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

			// Return a mock response
			w.WriteHeader(http.StatusOK)
			resp := []map[string]any{
				{
					"id": "mistral-tiny",
					"object": "model",
					"created": int64(1234567890),
					"owned_by": "mistral",
					"capabilities": map[string]any{
						"completion_chat": true,
						"completion_fim": false,
						"function_calling": true,
						"fine_tuning": false,
						"vision": false,
						"classification": false,
					},
				},
				{
					"id": "mistral-small",
					"object": "model",
					"created": int64(1234567891),
					"owned_by": "mistral",
					"capabilities": map[string]any{
						"completion_chat": true,
						"completion_fim": false,
						"function_calling": true,
						"fine_tuning": false,
						"vision": false,
						"classification": false,
					},
				},
			}
			require.NoError(t, json.NewEncoder(w).Encode(resp))
		}))
		defer server.Close()

		// Create a Mistral provider
		provider, err := NewProvider(WithAPIKey("test-api-key"), WithBaseURL(server.URL))
		require.NoError(t, err)

		// Call ListModels
		models, err := provider.ListModels(context.Background())
		require.NoError(t, err)
		require.NotNil(t, models)

		// Validate response
		require.Len(t, models, 2)
		assert.Equal(t, "mistral-tiny", models[0].ID)
		assert.Equal(t, "mistral-small", models[1].ID)
	})

	t.Run("should handle errors from the Mistral API", func(t *testing.T) {
		t.Parallel()

		// Mock server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": {"message": "invalid request", "type": "bad_request"}}`))
		}))
		defer server.Close()

		// Create a Mistral provider
		provider, err := NewProvider(WithAPIKey("test-api-key"), WithBaseURL(server.URL))
		require.NoError(t, err)

		// Call ListModels
		models, err := provider.ListModels(context.Background())
		assert.Error(t, err)
		assert.Nil(t, models)
	})
}

// TestMistralToolCalls tests tool/function calling with Mistral.
func TestMistralToolCalls(t *testing.T) {
	t.Parallel()

	t.Run("should handle tool calls in Generate", func(t *testing.T) {
		t.Parallel()

		// Mock server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/v1/chat/completions", r.URL.Path)

			// Validate request body
			var body map[string]any
			reqBody, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			require.NoError(t, json.Unmarshal(reqBody, &body))

			assert.Equal(t, "mistral-tiny", body["model"])

			// Return a mock response with tool calls
			w.WriteHeader(http.StatusOK)
			resp := map[string]any{
				"id":      "chatcmpl-test-id",
				"object":  "chat.completion",
				"created": 1234567890,
				"model":   "mistral-tiny",
				"choices": []map[string]any{
					{
						"index": 0,
						"message": map[string]any{
							"role": "assistant",
							"tool_calls": []map[string]any{
								{
									"id":   "call-test-id",
									"type": "function",
									"function": map[string]any{
										"name":      "get_weather",
										"arguments": `{"location":"Paris"}`,
									},
								},
							},
						},
						"finish_reason": "tool_calls",
					},
				},
				"usage": map[string]any{
					"prompt_tokens":     int64(50),
					"completion_tokens": int64(30),
					"total_tokens":      int64(80),
				},
			}
			require.NoError(t, json.NewEncoder(w).Encode(resp))
		}))
		defer server.Close()

		// Create a Mistral provider
		provider, err := NewProvider(WithAPIKey("test-api-key"), WithBaseURL(server.URL))
		require.NoError(t, err)

		// Create a language model
		lm := provider.LanguageModel("mistral-tiny")

		// Create a call with a tool
		call := fantasy.Call{
			Prompt: fantasy.Prompt{
				{
					Role: fantasy.MessageRoleUser,
					Content: []fantasy.MessagePart{
						fantasy.TextPart{Text: "What is the weather in Paris?"},
					},
				},
			},
			Tools: []fantasy.Tool{
				{
					Name: "get_weather",
					Description: "Get the weather for a location",
					InputSchema: &fantasy.Schema{
						Type: "object",
						Properties: map[string]*fantasy.Schema{
							"location": {
								Type: "string",
							},
						},
					},
				},
			},
		}

		// Call Generate
		resp, err := lm.Generate(context.Background(), call)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Validate response
		require.Len(t, resp.Content, 1)
		toolCall := resp.Content[0].(fantasy.ToolCallContent)
		assert.Equal(t, "get_weather", toolCall.ToolName)
		assert.Equal(t, `{"location":"Paris"}`, toolCall.Input)
	})
}
