package hub

import (
	"context"
	"errors"
	"testing"
)

func TestWrapSubscribeCallback(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		cb       interface{}
		wantErr  bool
		checkErr func(t *testing.T, err error)
	}{
		{
			name: "valid minimal callback",
			cb: func(ctx context.Context) error {
				return nil
			},
		},
		{
			name: "valid payload callback",
			cb: func(ctx context.Context, payload any) error {
				return nil
			},
		},
		{
			name: "valid topic+payload callback",
			cb: func(ctx context.Context, topic *Topic, payload any) error {
				return nil
			},
		},
		{
			name: "valid event callback",
			cb: func(ctx context.Context, e *Event) error {
				return nil
			},
		},
		{
			name:    "invalid non-function",
			cb:      "not a function",
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				if err == nil || err.Error() != "callback must be a function" {
					t.Errorf("expected 'callback must be a function' error, got %v", err)
				}
			},
		},
		{
			name: "invalid return count",
			cb: func(ctx context.Context) {
				// no return
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				if err == nil || err.Error() != "callback must return exactly one error value" {
					t.Errorf("expected return count error, got %v", err)
				}
			},
		},
		{
			name: "invalid first param",
			cb: func(notCtx string) error {
				return nil
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				if err == nil || err.Error() != "first parameter must be context.Context" {
					t.Errorf("expected context param error, got %v", err)
				}
			},
		},
		{
			name: "invalid topic param position",
			cb: func(ctx context.Context, payload any, topic *Topic) error {
				return nil
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				if err == nil || err.Error() != "second parameter must be *Topic when using three parameters" {
					t.Errorf("expected param count error, got %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := WrapSubscribeCallback(context.Background(), tt.cb)
			if (err != nil) != tt.wantErr {
				t.Errorf("WrapSubscribeCallback() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.checkErr != nil {
				tt.checkErr(t, err)
			}
		})
	}
}

func TestCallbackExecution(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testEvent := &Event{
		topic:   T("type=test"),
		payload: "test payload",
	}

	tests := []struct {
		name     string
		cb       interface{}
		expected error
	}{
		{
			name: "minimal callback",
			cb: func(ctx context.Context) error {
				return nil
			},
			expected: nil,
		},
		{
			name: "payload callback",
			cb: func(ctx context.Context, payload any) error {
				if payload != "test payload" {
					return errors.New("wrong payload")
				}
				return nil
			},
			expected: nil,
		},
		{
			name: "topic+payload callback",
			cb: func(ctx context.Context, topic *Topic, payload any) error {
				if topic.Get("type") != "test" {
					return errors.New("wrong topic")
				}
				if payload != "test payload" {
					return errors.New("wrong payload")
				}
				return nil
			},
			expected: nil,
		},
		{
			name: "event callback",
			cb: func(ctx context.Context, e *Event) error {
				if e != testEvent {
					return errors.New("wrong event")
				}
				return nil
			},
			expected: nil,
		},
		{
			name: "error return",
			cb: func(ctx context.Context) error {
				return errors.New("test error")
			},
			expected: errors.New("test error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy, err := WrapSubscribeCallback(ctx, tt.cb)
			if err != nil {
				t.Fatalf("WrapSubscribeCallback failed: %v", err)
			}

			err = proxy(ctx, testEvent)
			if (err == nil) != (tt.expected == nil) || (err != nil && err.Error() != tt.expected.Error()) {
				t.Errorf("proxy() = %v, want %v", err, tt.expected)
			}
		})
	}
}
