package common

type BusinessCode int
type Result struct {
	Code BusinessCode `json:"code"`
	Msg  string       `json:"msg"`
	Data any          `json:"data""`
}

func (r *Result) Success(data any) *Result {
	r.Code = 200
	r.Msg = "success"
	r.Data = data
	return r
}
func (r *Result) Fail(code BusinessCode, message string) *Result {
	r.Code = code
	r.Msg = message
	return r
}
