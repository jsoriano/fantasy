package mistral

// ModelInfo represents information about a Mistral model.
type ModelInfo struct {
	ID          string `json:"id"`
	Object      string `json:"object"`
	Created     int64  `json:"created"`
	OwnedBy     string `json:"owned_by"`
	Root        string `json:"root,omitempty"`
	Description string `json:"description,omitempty"`
	Capabilities struct {
		CompletionChat   bool `json:"completion_chat"`
		CompletionFIM    bool `json:"completion_fim"`
		FunctionCalling  bool `json:"function_calling"`
		FineTuning       bool `json:"fine_tuning"`
		Vision           bool `json:"vision"`
		Classification   bool `json:"classification"`
	} `json:"capabilities"`
}
