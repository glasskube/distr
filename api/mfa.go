package api

type SetupMFAResponse struct {
	Secret string `json:"secret"`
	QRCode string `json:"qrCode"`
}

type EnableMFARequest struct {
	Code string `json:"code"`
}

type VerifyMFARequest struct {
	Code string `json:"code"`
}

type DisableMFARequest struct {
	Password string `json:"password"`
}
