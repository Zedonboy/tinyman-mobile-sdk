package v1

type LibraryError struct {
	message string
}

func (e *LibraryError) Error() string {
	return e.message
}
