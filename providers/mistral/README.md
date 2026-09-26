# Mistral Provider

Mistral provider for the Fantasy library, supporting:
- Chat completions (non-streaming and streaming)
- Tool/function calling
- Model listing

## Usage

```go
import (
	"charm.land/fantasy"
	"charm.land/fantasy/providers/mistral"
)

// Create a Mistral provider
provider, err := mistral.NewProvider(
	mistral.WithAPIKey("your-api-key"),
)
if err != nil {
	// handle error
}

// Create a language model
lm := provider.LanguageModel("mistral-tiny")

// Generate a response
resp, err := lm.Generate(context.Background(), fantasy.Call{
	Prompt: fantasy.Prompt{
		fantasy.UserMessage("Hello!"),
	},
})
```

## Environment Variables

- `MISTRAL_API_KEY`: Mistral API key (required).

## Supported Models

- `mistral-tiny`
- `mistral-small`
- `mistral-medium`
- `mistral-large`
