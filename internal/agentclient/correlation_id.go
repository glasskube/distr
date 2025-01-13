package agentclient

import "net/http"

const CorrelationIDHeader = "X-Resource-Correlation-ID"

func getCorrelationID(r *http.Response) string {
	return r.Header.Get(CorrelationIDHeader)
}

func setCorrelationID(r *http.Request, value string) *http.Request {
	r.Header.Set(CorrelationIDHeader, value)
	return r
}
