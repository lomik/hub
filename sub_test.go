package hub

import (
	"context"
	"errors"
	"testing"
)

func TestSubCall(t *testing.T) {
	t.Run("event callback", func(t *testing.T) {
		s := &sub{
			handler: func(ctx context.Context, t *Topic, p any) error {
				return errors.New("test error")
			},
		}
		err := s.call(context.Background(), newEvent(nil, "type=test"))
		if err == nil || err.Error() != "test error" {
			t.Error("Expected test error from callback")
		}
	})

	t.Run("once", func(t *testing.T) {
		s := &sub{
			handler: func(ctx context.Context, t *Topic, p any) error {
				return errors.New("test error")
			},
			once: true,
		}
		err := s.call(context.Background(), newEvent(nil, "type=test"))
		if err == nil || err.Error() != "test error" {
			t.Error("Expected test error from callback")
		}

		err = s.call(context.Background(), newEvent(nil, "type=test"))
		if err != nil {
			t.Error("Expected no error on second call")
		}
	})

	t.Run("no callback", func(t *testing.T) {
		s := &sub{}
		err := s.call(context.Background(), newEvent(nil, "type=test"))
		if err != nil {
			t.Error("Expected nil error when no callbacks")
		}
	})
}
