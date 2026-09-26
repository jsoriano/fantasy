package providertests

import (
	"cmp"
	"net/http"
	"os"
	"testing"

	"charm.land/fantasy"
	"charm.land/fantasy/providers/mistral"
	"charm.land/x/vcr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMistralGenerate(t *testing.T) {
	r := vcr.NewRecorder(t)
	model := "mistral-small-latest"
	lm, err := mistralBuilder(model)(t, r)
	require.NoError(t, err)

	agent := fantasy.NewAgent(
		lm,
		fantasy.WithSystemPrompt("You are a helpful assistant"),
	)

	maxTokens := int64(4000)
	result, err := agent.Generate(t.Context(), fantasy.AgentCall{
		Prompt:          "Answer to the Ultimate Question of Life, the Universe, and Everything",
		MaxOutputTokens: &maxTokens,
	})
	require.NoError(t, err)

	got := result.Response.Content.Text()
	assert.NotEmpty(t, got, "should have a text response")
	t.Log(got)
}

func mistralBuilder(model string) builderFunc {
	return func(t *testing.T, r *vcr.Recorder) (fantasy.LanguageModel, error) {
		apiKey := cmp.Or(os.Getenv("FANTASY_MISTRAL_API_KEY"), "(missing)")
		provider, err := mistral.NewProvider(
			mistral.WithAPIKey(apiKey),
			mistral.WithHTTPClient(&http.Client{Transport: r}),
		)
		if err != nil {
			return nil, err
		}
		return provider.LanguageModel(t.Context(), model)
	}
}
