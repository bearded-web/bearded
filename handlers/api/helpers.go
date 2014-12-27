package api

type ErrorMsg struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type ApiError struct {
	Error ErrorMsg `json:"error"`
}

func NewApiError(code int, msg string) *ApiError {
	return &ApiError{
		Error: ErrorMsg{Code: code, Msg: msg}}
}
