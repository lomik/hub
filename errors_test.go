package hub

import (
	"errors"
	"testing"
)

func TestCastError(t *testing.T) {
	t.Run("normal error", func(t *testing.T) {
		origErr := errors.New("test error")
		ce := newCastError(origErr)

		if ce.orig != origErr {
			t.Errorf("Expected original error %v, got %v", origErr, ce.orig)
		}

		if ce.Error() != origErr.Error() {
			t.Errorf("Expected error message %q, got %q", origErr.Error(), ce.Error())
		}
	})

	t.Run("nil error", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic with nil error, but nothing happened")
			}
		}()

		// Это вызовет панику при вызове Error()
		ce := newCastError(nil)
		_ = ce.Error()
	})

	t.Run("error message propagation", func(t *testing.T) {
		testCases := []struct {
			msg       string
			expectMsg string
		}{
			{"invalid type", "invalid type"},
			{"conversion failed", "conversion failed"},
			{"", ""},
		}

		for _, tc := range testCases {
			t.Run(tc.msg, func(t *testing.T) {
				origErr := errors.New(tc.msg)
				ce := newCastError(origErr)

				if ce.Error() != tc.expectMsg {
					t.Errorf("Expected %q, got %q", tc.expectMsg, ce.Error())
				}
			})
		}
	})
}
