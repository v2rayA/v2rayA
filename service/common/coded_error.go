package common

type CodedError struct {
	Code   string
	Params map[string]interface{}
	Err    error
}

func (e *CodedError) Error() string {
	return e.Err.Error()
}

func (e *CodedError) Unwrap() error {
	return e.Err
}

func Coded(code string, err error, params map[string]interface{}) error {
	if err == nil {
		return nil
	}
	return &CodedError{
		Code:   code,
		Params: params,
		Err:    err,
	}
}
