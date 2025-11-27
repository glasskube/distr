package api

type CheckoutResponse struct {
	SessionID string `json:"sessionId"`
	URL       string `json:"url"`
}
