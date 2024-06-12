package eip712

// ForwardRequest represents the request structure
type ForwardRequest struct {
	From           string `json:"from"`
	To             string `json:"to"`
	Data           string `json:"data"`
	Value          string `json:"value"`
	Nonce          string `json:"nonce"`
	Gas            string `json:"gas"`
	ValidUntilTime string `json:"validUntilTime"`
}
