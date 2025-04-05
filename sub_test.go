package hub

import (
	"context"
	"errors"
	"testing"
)

func TestSubCall(t *testing.T) {
	t.Run("event callback", func(t *testing.T) {
		s := &sub{
			callbackEvent: func(ctx context.Context, e *Event) error {
				return errors.New("test error")
			},
		}
		err := s.call(context.Background(), E("type=test"))
		if err == nil || err.Error() != "test error" {
			t.Error("Expected test error from callback")
		}
	})

	t.Run("payload callback", func(t *testing.T) {
		s := &sub{
			callbackPayload: func(ctx context.Context, p any) error {
				return errors.New("payload error")
			},
		}
		err := s.call(context.Background(), E("type=test").WithPayload("data"))
		if err == nil || err.Error() != "payload error" {
			t.Error("Expected payload error from callback")
		}
	})

	t.Run("no callback", func(t *testing.T) {
		s := &sub{}
		err := s.call(context.Background(), E("type=test"))
		if err != nil {
			t.Error("Expected nil error when no callbacks")
		}
	})
}
