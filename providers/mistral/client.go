package mistral

import (
	"context"
	"fmt"

	"github.com/charmbracelet/openai-go"
	"github.com/charmbracelet/openai-go/option"
)

// newMistralClient creates a new Mistral API client.
func newMistralClient(apiKey, baseURL string) openai.Client {
	return openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL),
	)
}

// listMistralModels lists available Mistral models.
func listMistralModels(ctx context.Context, client openai.Client) ([]ModelInfo, error) {
	resp, err := client.Models.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list Mistral models: %w", err)
	}

	var models []ModelInfo
	for _, model := range resp.Data {
		models = append(models, ModelInfo{
			ID:      model.ID,
			Created: model.Created,
			OwnedBy: model.OwnedBy,
		})
	}

	return models, nil
}
