package hub

import (
	"context"
	"reflect"
	"testing"
)

func TestNewEvent(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"valid topic", []string{"type=alert"}, false},
		{"invalid topic", []string{"type"}, true},
		{"empty topic", []string{}, false},
		{"multiple pairs", []string{"a=1", "b=2"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewEvent(tt.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewEvent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Error("Expected non-nil event when no error")
			}
		})
	}
}

func TestE(t *testing.T) {
	t.Run("creates valid event", func(t *testing.T) {
		got := E("type=alert")
		if got == nil {
			t.Fatal("Expected non-nil event")
		}
		if got.Topic().String() != "type=alert" {
			t.Error("Incorrect topic setup")
		}
	})

	t.Run("panics on invalid input", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic with invalid input")
			}
		}()
		E("type")
	})
}

func TestEventGetters(t *testing.T) {
	t.Run("Topic() returns correct topic", func(t *testing.T) {
		ev := E("type=alert")
		if ev.Topic().String() != "type=alert" {
			t.Error("Topic() returned incorrect value")
		}
	})

	t.Run("Payload() returns nil by default", func(t *testing.T) {
		ev := E("type=alert")
		if ev.Payload() != nil {
			t.Error("Expected nil payload by default")
		}
	})

	t.Run("hasOnFinish() returns false initially", func(t *testing.T) {
		ev := E("type=alert")
		if ev.hasOnFinish() {
			t.Error("Expected no finish handlers initially")
		}
	})
}

func TestEventWithMethods(t *testing.T) {
	base := E("type=alert")

	t.Run("WithPayload()", func(t *testing.T) {
		ev := base.WithPayload("test")
		if ev.Payload() != "test" {
			t.Error("Payload not set correctly")
		}
		if base.Payload() != nil {
			t.Error("Original event modified")
		}
	})

	t.Run("WithTopic()", func(t *testing.T) {
		newTopic := T("type=update")
		ev := base.WithTopic(newTopic)
		if ev.Topic() != newTopic {
			t.Error("Topic not set correctly")
		}
		if base.Topic().String() != "type=alert" {
			t.Error("Original event modified")
		}
	})

	t.Run("WithWait()", func(t *testing.T) {
		ev := base.WithWait(true)
		if !ev.wait {
			t.Error("Wait flag not set")
		}
		if base.wait {
			t.Error("Original event modified")
		}
	})

	t.Run("WithSync()", func(t *testing.T) {
		ev := base.WithSync(true)
		if !ev.sync {
			t.Error("Sync flag not set")
		}
		if base.sync {
			t.Error("Original event modified")
		}
	})

	t.Run("WithOnFinish() adds callback", func(t *testing.T) {
		called := false
		cb := func(context.Context, *Event) { called = true }

		ev := base.WithOnFinish(cb)
		ev.finish(context.Background())

		if !called {
			t.Error("Callback not executed")
		}
		if base.hasOnFinish() {
			t.Error("Original event modified")
		}
	})
}

func TestEventFinish(t *testing.T) {
	t.Run("executes multiple callbacks in order", func(t *testing.T) {
		var calls []int
		cb1 := func(context.Context, *Event) { calls = append(calls, 1) }
		cb2 := func(context.Context, *Event) { calls = append(calls, 2) }

		ev := E("type=test").
			WithOnFinish(cb1).
			WithOnFinish(cb2)

		ev.finish(context.Background())

		if !reflect.DeepEqual(calls, []int{1, 2}) {
			t.Error("Callbacks not executed in order")
		}
	})

	t.Run("passes correct context and event", func(t *testing.T) {
		testCtx := context.WithValue(context.Background(), "key", "value")
		var gotCtx context.Context
		var gotEv *Event

		cb := func(ctx context.Context, ev *Event) {
			gotCtx = ctx
			gotEv = ev
		}

		ev := E("type=test").WithOnFinish(cb)
		ev.finish(testCtx)

		if gotCtx != testCtx {
			t.Error("Wrong context passed to callback")
		}
		if gotEv != ev {
			t.Error("Wrong event passed to callback")
		}
	})

	t.Run("handles nil callback", func(t *testing.T) {
		// Should not panic
		ev := E("type=test").WithOnFinish(nil)
		ev.finish(context.Background())
	})
}

func TestEventClone(t *testing.T) {
	t.Run("creates independent copy", func(t *testing.T) {
		original := E("type=alert").
			WithPayload("test").
			WithWait(true).
			WithSync(true).
			WithOnFinish(func(context.Context, *Event) {})

		clone := original.clone()

		// Verify all fields were copied
		if clone.Topic().String() != original.Topic().String() {
			t.Error("Topic not cloned correctly")
		}
		if clone.Payload() != original.Payload() {
			t.Error("Payload not cloned correctly")
		}
		if clone.wait != original.wait {
			t.Error("Wait flag not cloned correctly")
		}
		if clone.sync != original.sync {
			t.Error("Sync flag not cloned correctly")
		}
		if len(clone.onFinish) != len(original.onFinish) {
			t.Error("Callbacks not cloned correctly")
		}

		// Verify it's a new instance
		if clone == original {
			t.Error("Clone returned same instance")
		}
	})
}

func TestEventCombinations(t *testing.T) {
	t.Run("chained With methods", func(t *testing.T) {
		ev := E("type=alert").
			WithPayload("data").
			WithTopic(T("type=update")).
			WithWait(true).
			WithSync(true).
			WithOnFinish(func(context.Context, *Event) {})

		if ev.Payload() != "data" {
			t.Error("Payload not set")
		}
		if ev.Topic().String() != "type=update" {
			t.Error("Topic not updated")
		}
		if !ev.wait {
			t.Error("Wait flag not set")
		}
		if !ev.sync {
			t.Error("Sync flag not set")
		}
		if !ev.hasOnFinish() {
			t.Error("Callback not set")
		}
	})
}
