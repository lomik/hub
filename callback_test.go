package hub

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestWrapSubscribeCallback(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name        string
		cb          interface{}
		wantErr     bool
		errContains string
	}{
		// Invalid callbacks
		{
			name:        "not a function",
			cb:          "not a function",
			wantErr:     true,
			errContains: "callback must be a function",
		},
		{
			name: "invalid return type",
			cb: func(ctx context.Context) string {
				return "not an error"
			},
			wantErr:     true,
			errContains: "callback must return exactly one error value",
		},
		{
			name: "no parameters",
			cb: func() error {
				return nil
			},
			wantErr:     true,
			errContains: "callback must have 1-2 parameters",
		},
		{
			name: "wrong first parameter",
			cb: func(notCtx string) error {
				return nil
			},
			wantErr:     true,
			errContains: "first parameter must be context.Context",
		},
		{
			name: "unsupported type",
			cb: func(ctx context.Context, ch chan int) error {
				return nil
			},
			wantErr:     true,
			errContains: "unsupported parameter type",
		},

		// Valid callbacks
		{
			name: "minimal callback",
			cb: func(ctx context.Context) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "event callback",
			cb: func(ctx context.Context, e *Event) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "generic any callback",
			cb: func(ctx context.Context, a any) error {
				return nil
			},
			wantErr: false,
		},
	}

	// Add tests for all supported types
	supportedTypes := []struct {
		name string
		cb   interface{}
	}{
		{"string", func(ctx context.Context, s string) error { return nil }},
		{"int", func(ctx context.Context, i int) error { return nil }},
		{"int8", func(ctx context.Context, i int8) error { return nil }},
		{"int16", func(ctx context.Context, i int16) error { return nil }},
		{"int32", func(ctx context.Context, i int32) error { return nil }},
		{"int64", func(ctx context.Context, i int64) error { return nil }},
		{"uint", func(ctx context.Context, i uint) error { return nil }},
		{"uint8", func(ctx context.Context, i uint8) error { return nil }},
		{"uint16", func(ctx context.Context, i uint16) error { return nil }},
		{"uint32", func(ctx context.Context, i uint32) error { return nil }},
		{"uint64", func(ctx context.Context, i uint64) error { return nil }},
		{"float32", func(ctx context.Context, f float32) error { return nil }},
		{"float64", func(ctx context.Context, f float64) error { return nil }},
		{"bool", func(ctx context.Context, b bool) error { return nil }},
		{"time.Time", func(ctx context.Context, tm time.Time) error { return nil }},
		{"time.Duration", func(ctx context.Context, d time.Duration) error { return nil }},
		{"[]string", func(ctx context.Context, s []string) error { return nil }},
		{"map[string]any", func(ctx context.Context, m map[string]any) error { return nil }},
	}

	for _, typ := range supportedTypes {
		tests = append(tests, struct {
			name        string
			cb          interface{}
			wantErr     bool
			errContains string
		}{
			name:    "supported type: " + typ.name,
			cb:      typ.cb,
			wantErr: false,
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := wrapSubscribeCallback(ctx, tt.cb)
			if (err != nil) != tt.wantErr {
				t.Errorf("wrapSubscribeCallback() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" && !contains(err.Error(), tt.errContains) {
				t.Errorf("Expected error to contain %q, got %v", tt.errContains, err)
			}
		})
	}
}

func TestWrappedCallbackExecution(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testCases := []struct {
		name        string
		payload     interface{}
		cb          interface{}
		expectError bool
	}{
		{
			name:    "string direct match",
			payload: "test",
			cb: func(ctx context.Context, s string) error {
				if s != "test" {
					return errors.New("unexpected value")
				}
				return nil
			},
		},
		{
			name:    "string conversion",
			payload: 123,
			cb: func(ctx context.Context, s string) error {
				if s != "123" {
					return errors.New("conversion failed")
				}
				return nil
			},
		},
		{
			name:    "int direct match",
			payload: 42,
			cb: func(ctx context.Context, i int) error {
				if i != 42 {
					return errors.New("unexpected value")
				}
				return nil
			},
		},
		{
			name:    "int conversion from string",
			payload: "42",
			cb: func(ctx context.Context, i int) error {
				if i != 42 {
					return errors.New("conversion failed")
				}
				return nil
			},
		},
		{
			name:    "time.Time conversion",
			payload: "2023-01-01T00:00:00Z",
			cb: func(ctx context.Context, tm time.Time) error {
				if tm.Year() != 2023 {
					return errors.New("conversion failed")
				}
				return nil
			},
		},
		{
			name: "generic any callback",
			payload: struct {
				Field string
			}{Field: "test"},
			cb: func(ctx context.Context, a any) error {
				if reflect.TypeOf(a).Kind() != reflect.Struct {
					return errors.New("unexpected type")
				}
				return nil
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := &Event{
				topic:   T("type=test"),
				payload: tc.payload,
			}

			proxy, err := wrapSubscribeCallback(ctx, tc.cb)
			if err != nil {
				t.Fatalf("wrapSubscribeCallback failed: %v", err)
			}

			err = proxy(ctx, event)
			if (err != nil) != tc.expectError {
				t.Errorf("Unexpected error state: got %v, want error=%v", err, tc.expectError)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}
