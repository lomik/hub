package hub

// CastError represents an error that occurs during type casting.
type CastError struct {
	orig error
}

// Error implements the error interface for CastError.
func (e *CastError) Error() string {
	return e.orig.Error()
}

// newCastError creates a new instance of CastError.
func newCastError(orig error) *CastError {
	return &CastError{
		orig: orig,
	}
}
