package mistral

import (
	"strings"

	"charm.land/fantasy"
	"github.com/charmbracelet/openai-go"
)

// DefaultToPrompt converts a fantasy prompt to Mistral format.
func DefaultToPrompt(prompt fantasy.Prompt, _, _ string) ([]openai.ChatCompletionMessageParamUnion, []fantasy.CallWarning) {
	var messages []openai.ChatCompletionMessageParamUnion
	var warnings []fantasy.CallWarning

	for _, msg := range prompt {
		switch msg.Role {
		case fantasy.MessageRoleSystem:
			var systemPromptParts []string
			for _, c := range msg.Content {
				if c.GetType() != fantasy.ContentTypeText {
					warnings = append(warnings, fantasy.CallWarning{
						Type:    fantasy.CallWarningTypeOther,
						Message: "system prompt can only have text content",
					})
					continue
				}
				textPart, ok := fantasy.AsContentType[fantasy.TextPart](c)
				if !ok {
					warnings = append(warnings, fantasy.CallWarning{
						Type:    fantasy.CallWarningTypeOther,
						Message: "system prompt text part does not have the right type",
					})
					continue
				}
				text := textPart.Text
				if strings.TrimSpace(text) != "" {
					systemPromptParts = append(systemPromptParts, textPart.Text)
				}
			}
			if len(systemPromptParts) == 0 {
				warnings = append(warnings, fantasy.CallWarning{
					Type:    fantasy.CallWarningTypeOther,
					Message: "system prompt has no text parts",
				})
				continue
			}
			messages = append(messages, openai.SystemMessage(strings.Join(systemPromptParts, "\n")))
		case fantasy.MessageRoleUser:
			// simple user message just text content
			if len(msg.Content) == 1 && msg.Content[0].GetType() == fantasy.ContentTypeText {
				textPart, ok := fantasy.AsContentType[fantasy.TextPart](msg.Content[0])
				if !ok {
					warnings = append(warnings, fantasy.CallWarning{
						Type:    fantasy.CallWarningTypeOther,
						Message: "user message text part does not have the right type",
					})
					continue
				}
				messages = append(messages, openai.UserMessage(textPart.Text))
				continue
			}

			// complex user message with multiple parts
			var userPromptParts []string
			for _, c := range msg.Content {
				if c.GetType() != fantasy.ContentTypeText {
					warnings = append(warnings, fantasy.CallWarning{
						Type:    fantasy.CallWarningTypeOther,
						Message: "user message can only have text content",
					})
					continue
				}
				textPart, ok := fantasy.AsContentType[fantasy.TextPart](c)
				if !ok {
					warnings = append(warnings, fantasy.CallWarning{
						Type:    fantasy.CallWarningTypeOther,
						Message: "user message text part does not have the right type",
					})
					continue
				}
				text := textPart.Text
				if strings.TrimSpace(text) != "" {
					userPromptParts = append(userPromptParts, textPart.Text)
				}
			}
			if len(userPromptParts) == 0 {
				warnings = append(warnings, fantasy.CallWarning{
					Type:    fantasy.CallWarningTypeOther,
					Message: "user message has no text parts",
				})
				continue
			}
			messages = append(messages, openai.UserMessage(strings.Join(userPromptParts, "\n")))
		case fantasy.MessageRoleAssistant:
			var assistantParts []string
			for _, c := range msg.Content {
				if c.GetType() != fantasy.ContentTypeText {
					warnings = append(warnings, fantasy.CallWarning{
						Type:    fantasy.CallWarningTypeOther,
						Message: "assistant message can only have text content",
					})
					continue
				}
				textPart, ok := fantasy.AsContentType[fantasy.TextPart](c)
				if !ok {
					warnings = append(warnings, fantasy.CallWarning{
						Type:    fantasy.CallWarningTypeOther,
						Message: "assistant message text part does not have the right type",
					})
					continue
				}
				text := textPart.Text
				if strings.TrimSpace(text) != "" {
					assistantParts = append(assistantParts, textPart.Text)
				}
			}
			if len(assistantParts) == 0 {
				warnings = append(warnings, fantasy.CallWarning{
					Type:    fantasy.CallWarningTypeOther,
					Message: "assistant message has no text parts",
				})
				continue
			}
			messages = append(messages, openai.AssistantMessage(strings.Join(assistantParts, "\n")))
		}
	}

	return messages, warnings
}