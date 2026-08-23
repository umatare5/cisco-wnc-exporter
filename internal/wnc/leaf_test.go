package wnc

// ptr returns a pointer to v. The SDK types a leaf as a pointer so that a leaf the controller
// omitted stays distinguishable from one it sent at its zero value, and a fixture setting such
// a leaf has no addressable constant to take.
func ptr[T any](v T) *T { return &v }
