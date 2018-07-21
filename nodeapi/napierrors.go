package nodeapi

type NapiKeyError struct {
	message string
}

func (r NapiKeyError) Error() string {
	return r.message
}

func NewNapiKeyError() *NapiKeyError {
	return &NapiKeyError{message: "Node API key error"}
}

type NapiRangeError struct {
	message string
}

func (r NapiRangeError) Error() string {
	return r.message
}

func NewNapiRangeError() *NapiRangeError {
	return &NapiRangeError{message: "Node API range error"}
}

type NapiCallerError struct {
	message string
}

func (r NapiCallerError) Error() string {
	return r.message
}

func NewNapiCallerError() *NapiCallerError {
	return &NapiCallerError{message: "Node API Caller error"}
}

type NapiBusyError struct {
	message string
}

func (r NapiBusyError) Error() string {
	return r.message
}

func NewNapiBusyError() *NapiBusyError {
	return &NapiBusyError{message: "Node API Busy error"}
}
