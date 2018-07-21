package fingertable

type FTFindError struct {
	message string
}

func (r FTFindError) Error() string {
	return r.message
}

func NewFTFindError() *FTFindError {
	return &FTFindError{message: "FingerTable cannot find node for given key"}
}
