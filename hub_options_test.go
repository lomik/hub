package hub

import (
	"context"
	"errors"
	"testing"

	"github.com/spf13/cast"
)

func TestCustomHandlerConversion(t *testing.T) {
	t.Parallel()

	// Test converter that handles func(string) error
	stringConverter := func(ctx context.Context, cb any) (Handler, error) {
		if fn, ok := cb.(func(string) error); ok {
			return func(ctx context.Context, e *Event) error {
				s, err := cast.ToStringE(e.Payload())
				if err != nil {
					return err
				}
				return fn(s)
			}, nil
		}
		return nil, errors.New("unsupported type")
	}

	// Test converter that handles func(int) error
	intConverter := func(ctx context.Context, cb any) (Handler, error) {
		if fn, ok := cb.(func(int) error); ok {
			return func(ctx context.Context, e *Event) error {
				i, err := cast.ToIntE(e.Payload())
				if err != nil {
					return err
				}
				return fn(i)
			}, nil
		}
		return nil, errors.New("unsupported type")
	}

	tests := []struct {
		name        string
		converter   HubOption
		cb          interface{}
		payload     interface{}
		wantErr     bool
		expectError string
	}{
		{
			name:      "successful string conversion",
			converter: ToHandler(stringConverter),
			cb: func(s string) error {
				if s != "test" {
					return errors.New("unexpected value")
				}
				return nil
			},
			payload: "test",
		},
		{
			name:      "successful int conversion",
			converter: ToHandler(intConverter),
			cb: func(i int) error {
				if i != 42 {
					return errors.New("unexpected value")
				}
				return nil
			},
			payload: 42,
		},
		{
			name:        "unsupported callback type",
			converter:   ToHandler(intConverter),
			cb:          func(s string) error { return nil }, // Only int converter registered
			payload:     "test",
			wantErr:     true,
			expectError: "unsupported type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := New(tt.converter)

			// Test conversion
			handler, err := h.ToHandler(context.Background(), tt.cb)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				// Test execution
				event := &Event{
					topic:   T("type=test"),
					payload: tt.payload,
				}

				err = handler(context.Background(), event)
				if tt.expectError != "" && (err == nil || err.Error() != tt.expectError) {
					t.Errorf("handler execution error = %v, expectError %v", err, tt.expectError)
				}
			} else if tt.expectError != "" && !contains(err.Error(), tt.expectError) {
				t.Errorf("ToHandler() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestMultipleConverters(t *testing.T) {
	t.Parallel()

	// First converter handles func(string)
	stringConverter := ToHandler(func(ctx context.Context, cb any) (Handler, error) {
		if fn, ok := cb.(func(string) error); ok {
			return func(ctx context.Context, e *Event) error {
				return fn(cast.ToString(e.Payload()))
			}, nil
		}
		return nil, nil
	})

	// Second converter handles func(int)
	intConverter := ToHandler(func(ctx context.Context, cb any) (Handler, error) {
		if fn, ok := cb.(func(int) error); ok {
			return func(ctx context.Context, e *Event) error {
				return fn(cast.ToInt(e.Payload()))
			}, nil
		}
		return nil, nil
	})

	h := New(stringConverter, intConverter)

	t.Run("string handler via first converter", func(t *testing.T) {
		called := false
		_, err := h.Subscribe(context.Background(), T("type=test"), func(s string) error {
			if s != "test" {
				return errors.New("unexpected value")
			}
			called = true
			return nil
		})

		if err != nil {
			t.Fatalf("Subscribe failed: %v", err)
		}

		h.Publish(context.Background(), T("type=test"), "test", Sync(true))
		if !called {
			t.Error("string handler was not called")
		}
	})

	t.Run("int handler via second converter", func(t *testing.T) {
		called := false
		_, err := h.Subscribe(context.Background(), T("type=test"), func(i int) error {
			if i != 42 {
				return errors.New("unexpected value")
			}
			called = true
			return nil
		})

		if err != nil {
			t.Fatalf("Subscribe failed: %v", err)
		}

		h.Publish(context.Background(), T("type=test"), 42, Sync(true))
		if !called {
			t.Error("int handler was not called")
		}
	})

	t.Run("unsupported type", func(t *testing.T) {
		_, err := h.Subscribe(context.Background(), T("type=test"), func(b bool) error {
			return nil
		})

		if err == nil {
			t.Error("expected error for unsupported type")
		}
	})
}
