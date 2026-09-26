package mistral

import (
	"context"

	"charm.land/fantasy"
)

// Name is the name of the Mistral provider.
const Name = "mistral"

// Provider is the Mistral provider implementation.
type Provider struct {
	apiKey  string
	baseURL string
}

// ProviderOption is a function that configures the Mistral provider.
type ProviderOption func(*Provider)

// WithAPIKey sets the API key for the Mistral provider.
func WithAPIKey(apiKey string) ProviderOption {
	return func(p *Provider) {
		p.apiKey = apiKey
	}
}

// WithBaseURL sets the base URL for the Mistral provider.
func WithBaseURL(baseURL string) ProviderOption {
	return func(p *Provider) {
		p.baseURL = baseURL
	}
}

// NewProvider creates a new Mistral provider.
func NewProvider(opts ...ProviderOption) (*Provider, error) {
	p := &Provider{
		apiKey:  "",
		baseURL: "https://api.mistral.ai",
	}

	for _, opt := range opts {
		opt(p)
	}

	if p.apiKey == "" {
		p.apiKey = getAPIKeyFromEnv()
	}

	return p, nil
}

// Name implements fantasy.Provider.
func (p *Provider) Name() string {
	return Name
}

// LanguageModel implements fantasy.Provider.
func (p *Provider) LanguageModel(_ context.Context, modelID string) (fantasy.LanguageModel, error) {
	client := newMistralClient(p.apiKey, p.baseURL)

	return newMistralLanguageModel(modelID, client), nil
}

// ListModels implements fantasy.Provider.
func (p *Provider) ListModels(ctx context.Context) ([]ModelInfo, error) {
	client := newMistralClient(p.apiKey, p.baseURL)

	return listMistralModels(ctx, client)
}

// getAPIKeyFromEnv returns the Mistral API key from the environment.
func getAPIKeyFromEnv() string {
	return ""
}
