package collector

import "testing"

// ptr returns a pointer to v. The SDK types a leaf as a pointer so that a leaf the controller
// omitted stays distinguishable from one it sent at its zero value, and a fixture setting such
// a leaf has no addressable constant to take.
func ptr[T any](v T) *T { return &v }

// wantLeaf fails unless got carries exactly what want does, absence included. Comparing the
// two pointers directly would compare addresses, and absence is the case these accessors exist
// to report.
func wantLeaf[T comparable](t *testing.T, name string, got, want *T) {
	t.Helper()

	switch {
	case got == nil && want == nil:
	case got == nil:
		t.Errorf("%s() = nil, want %v", name, *want)
	case want == nil:
		t.Errorf("%s() = %v, want nil", name, *got)
	case *got != *want:
		t.Errorf("%s() = %v, want %v", name, *got, *want)
	}
}
