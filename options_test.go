package hub

import (
	"context"
	"testing"
)

func TestOnce(t *testing.T) {
	t.Run("sets once flag true", func(t *testing.T) {
		opt := Once(true)
		s := &sub{}
		opt.modifySub(context.Background(), s)
		if !s.once {
			t.Error("Once(true) didn't set sub.once to true")
		}
	})

	t.Run("sets once flag false", func(t *testing.T) {
		opt := Once(false)
		s := &sub{once: true}
		opt.modifySub(context.Background(), s)
		if s.once {
			t.Error("Once(false) didn't set sub.once to false")
		}
	})
}

func TestSync(t *testing.T) {
	t.Run("enables sync mode", func(t *testing.T) {
		opt := Sync(true)
		e := &Event{}
		modified := opt.modifyEvent(context.Background(), e)
		if !modified.sync {
			t.Error("Sync(true) didn't enable sync mode")
		}
	})

	t.Run("disables sync mode", func(t *testing.T) {
		opt := Sync(false)
		e := &Event{sync: true}
		modified := opt.modifyEvent(context.Background(), e)
		if modified.sync {
			t.Error("Sync(false) didn't disable sync mode")
		}
	})
}

func TestWait(t *testing.T) {
	t.Run("enables wait mode", func(t *testing.T) {
		opt := Wait(true)
		e := &Event{}
		modified := opt.modifyEvent(context.Background(), e)
		if !modified.wait {
			t.Error("Wait(true) didn't enable wait mode")
		}
	})

	t.Run("disables wait mode", func(t *testing.T) {
		opt := Wait(false)
		e := &Event{wait: true}
		modified := opt.modifyEvent(context.Background(), e)
		if modified.wait {
			t.Error("Wait(false) didn't disable wait mode")
		}
	})
}

func TestOnFinish(t *testing.T) {
	t.Run("sets callback function", func(t *testing.T) {
		called := false
		cb := func(ctx context.Context, ev *Event) {
			called = true
		}

		opt := OnFinish(cb)
		e := &Event{}
		modified := opt.modifyEvent(context.Background(), e)

		// Verify callback was set
		if len(modified.onFinish) != 1 {
			t.Fatal("Callback not set")
		}

		// Test callback execution
		modified.onFinish[0](context.Background(), modified)
		if !called {
			t.Error("Callback function not executed properly")
		}
	})

	t.Run("nil callback", func(t *testing.T) {
		opt := OnFinish(nil)
		e := &Event{}
		modified := opt.modifyEvent(context.Background(), e)
		if len(modified.onFinish) != 0 {
			t.Error("Nil callback should not be added")
		}
	})
}

func TestOptionChaining(t *testing.T) {
	t.Run("multiple publish options", func(t *testing.T) {
		e := &Event{}
		opts := []PublishOption{
			Sync(true),
			Wait(true),
			OnFinish(func(ctx context.Context, ev *Event) {}),
		}

		for _, opt := range opts {
			e = opt.modifyEvent(context.Background(), e)
		}

		if !e.sync || !e.wait || len(e.onFinish) != 1 {
			t.Error("Failed to apply multiple options correctly")
		}
	})
}

func TestOptionImmutable(t *testing.T) {
	t.Run("original event not modified", func(t *testing.T) {
		original := &Event{}
		opt := Sync(true)
		modified := opt.modifyEvent(context.Background(), original)

		if original.sync {
			t.Error("Original event was modified")
		}
		if !modified.sync {
			t.Error("Modified event not updated")
		}
		if original == modified {
			t.Error("Should return new Event instance")
		}
	})
}
