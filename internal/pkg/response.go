package pkg

// ErrorEnvelope is the error response shape used by every endpoint (see
// docs/api-spec.md): {"error": {"code": "...", "message": "..."}}.
type ErrorEnvelope struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewErrorEnvelope(code, message string) ErrorEnvelope {
	return ErrorEnvelope{Error: ErrorBody{Code: code, Message: message}}
}
