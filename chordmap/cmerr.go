package chordmap

type CMRangeError struct {
	message string
}

func (r CMRangeError) Error() string {
	return r.message
}

func NewCMRangeError() *CMRangeError {
	return &CMRangeError{message: "ChordMap range error"}
}

type CMKeyError struct {
	message string
}

func (r CMKeyError) Error() string {
	return r.message
}

func NewCMKeyError() *CMKeyError {
	return &CMKeyError{message: "ChordMap does not contain key"}
}

type CMConvError struct {
	message string
}

func (r CMConvError) Error() string {
	return r.message
}

func NewCMConvError() *CMConvError {
	return &CMConvError{message: "ShaStr conversion error"}
}
