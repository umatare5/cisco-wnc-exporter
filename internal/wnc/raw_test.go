package wnc

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	wnc "github.com/umatare5/cisco-ios-xe-wireless-go"
)

// stubGetter answers one raw read with a canned body or error.
type stubGetter struct {
	body string
	err  error
	path string
}

func (s *stubGetter) GetData(_ context.Context, path string, _ ...wnc.GetOption) ([]byte, error) {
	s.path = path
	if s.err != nil {
		return nil, s.err
	}
	return []byte(s.body), nil
}

// TestRawValue_RejectsAWrongEnvelopeKey is the reason this read decodes without struct
// tags. A body whose envelope key is not the node the path names must fail loudly: with
// a tag-shaped decode it yields an empty value, a nil error and HTTP 200, so the family
// disappears while the refresh reports success. Counting the keys alone does not catch
// it, because the wrong body below carries exactly one.
func TestRawValue_RejectsAWrongEnvelopeKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{
			name: "the key the path names",
			body: `{"Cisco-IOS-XE-wireless-client-global-oper:co-client-del-reason":{"ap-delete":1}}`,
		},
		{
			name:    "another node of the same module",
			body:    `{"Cisco-IOS-XE-wireless-client-global-oper:client-stats":{"ap-delete":1}}`,
			wantErr: true,
		},
		{
			name:    "the right node without a module prefix",
			body:    `{"co-client-del-reason":{"ap-delete":1}}`,
			wantErr: true,
		},
		{
			name: "two keys",
			body: `{"Cisco-IOS-XE-wireless-client-global-oper:co-client-del-reason":{},` +
				`"Cisco-IOS-XE-wireless-client-global-oper:client-stats":{}}`,
			wantErr: true,
		},
		{
			name:    "no key at all",
			body:    `{}`,
			wantErr: true,
		},
		{
			name:    "not an object",
			body:    `[]`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			getter := &stubGetter{body: tt.body}
			value, present, err := rawValue[map[string]json.RawMessage](
				context.Background(), getter, routeCoClientDelReason,
			)

			if (err != nil) != tt.wantErr {
				t.Fatalf("rawValue() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if present {
					t.Error("rawValue() reported the value as present alongside an error")
				}
				return
			}
			if !present || len(value) == 0 {
				t.Errorf("rawValue() present = %v with %d leaves, want a decoded container", present, len(value))
			}
		})
	}
}

// TestRawValue_ReadsThePathItsRouteNames pins the URL this read builds. The method is no
// longer a parameter, so the path is the whole of what is under this file's control, and it
// uses no SDK route constant, so nothing else asserts it.
func TestRawValue_ReadsThePathItsRouteNames(t *testing.T) {
	t.Parallel()

	getter := &stubGetter{body: `{"Cisco-IOS-XE-device-hardware-oper:boot-time":"2026-01-01T00:00:00+00:00"}`}
	if _, _, err := rawValue[string](context.Background(), getter, routeControllerBootTime); err != nil {
		t.Fatalf("rawValue() error = %v, want nil", err)
	}

	if want := restconfDataPath + routeControllerBootTime; getter.path != want {
		t.Errorf("rawValue() read %q, want %q", getter.path, want)
	}
}

// TestRawValue_EmptyBodyIsAbsenceAndAnErrorIsNot separates the two states an operator
// has to act on differently. This platform answers a container it does not carry with
// no body at all, which is the same accounting a typed accessor gives an empty
// response; a path this exporter got wrong answers 404, and that must stay a failure so
// it cannot hide behind the same silence.
func TestRawValue_EmptyBodyIsAbsenceAndAnErrorIsNot(t *testing.T) {
	t.Parallel()

	value, present, err := rawValue[string](context.Background(), &stubGetter{body: ""}, routeControllerBootTime)
	if err != nil {
		t.Errorf("rawValue() error = %v for an empty body, want nil", err)
	}
	if present || value != "" {
		t.Errorf("rawValue() present = %v value = %q for an empty body, want absence", present, value)
	}

	failing := &stubGetter{err: errors.New("404 not found")}
	if _, present, err := rawValue[string](context.Background(), failing, routeControllerBootTime); err == nil {
		t.Error("rawValue() error = nil for a failed request, want the error")
	} else if present {
		t.Error("rawValue() reported the value as present alongside a failed request")
	}
}

// TestNumericLeaves_SkipsOneLeafRatherThanTheContainer pins the per-leaf decode. This
// controller writes some 64-bit counters as JSON strings, so a leaf has to be read
// through a decoder that accepts both; and decoding the whole container that way at
// once would abandon every leaf after the first one that is neither, which for this
// read would drop hundreds of counters because one leaf changed shape.
func TestNumericLeaves_SkipsOneLeafRatherThanTheContainer(t *testing.T) {
	t.Parallel()

	var leaves map[string]json.RawMessage
	body := `{"number":1,"numeric-string":"697517218065","word":"abc","boolean":true,"nested":{"x":1}}`
	if err := json.Unmarshal([]byte(body), &leaves); err != nil {
		t.Fatalf("Unmarshal() error = %v, want nil", err)
	}

	values := numericLeaves(leaves, "test")

	want := map[string]float64{"number": 1, "numeric-string": 697517218065}
	if len(values) != len(want) {
		t.Errorf("numericLeaves() kept %d leaves, want %d: %v", len(values), len(want), values)
	}
	for leaf, wantValue := range want {
		if got, ok := values[leaf]; !ok || got != wantValue {
			t.Errorf("numericLeaves()[%s] = %v (present %v), want %v", leaf, got, ok, wantValue)
		}
	}
	for _, leaf := range []string{"word", "boolean", "nested"} {
		if _, ok := values[leaf]; ok {
			t.Errorf("numericLeaves() kept %s, want it skipped", leaf)
		}
	}
}
