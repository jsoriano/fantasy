package mistral

import (
	"encoding/json"

	"charm.land/fantasy"
)

// Global type identifiers for Mistral-specific provider data.
const (
	TypeProviderOptions  = Name + ".options"
	TypeProviderMetadata = Name + ".metadata"
)

// ProviderOptions represents Mistral-specific provider options.
type ProviderOptions struct {
	APIKey string `json:"api_key,omitempty"`
}

// Options implements the ProviderOptions interface.
func (*ProviderOptions) Options() {}

// MarshalJSON implements json.Marshaler.
func (o ProviderOptions) MarshalJSON() ([]byte, error) {
	type plain ProviderOptions // avoid infinite recursion
	return fantasy.MarshalProviderType(TypeProviderOptions, plain(o))
}

// UnmarshalJSON implements json.Unmarshaler.
func (o *ProviderOptions) UnmarshalJSON(data []byte) error {
	type plain ProviderOptions // avoid infinite recursion
	return fantasy.UnmarshalProviderType(data, &plain{})
}

// ProviderMetadata represents Mistral-specific provider metadata.
type ProviderMetadata struct{}

// Options implements the ProviderOptions interface.
func (*ProviderMetadata) Options() {}

// MarshalJSON implements json.Marshaler.
func (m ProviderMetadata) MarshalJSON() ([]byte, error) {
	type plain ProviderMetadata // avoid infinite recursion
	return fantasy.MarshalProviderType(TypeProviderMetadata, plain(m))
}

// UnmarshalJSON implements json.Unmarshaler.
func (m *ProviderMetadata) UnmarshalJSON(data []byte) error {
	type plain ProviderMetadata // avoid infinite recursion
	return fantasy.UnmarshalProviderType(data, &plain{})
}

// Register Mistral provider-specific types with the global registry.
func init() {
	fantasy.RegisterProviderType(TypeProviderOptions, func(data []byte) (fantasy.ProviderOptionsData, error) {
		var v ProviderOptions
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		return &v, nil
	})

	fantasy.RegisterProviderType(TypeProviderMetadata, func(data []byte) (fantasy.ProviderOptionsData, error) {
		var v ProviderMetadata
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		return &v, nil
	})
}