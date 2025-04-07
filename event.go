package hub

import (
	"context"
)

// Event represents a message sent to a specific topic in the event hub.
// It contains the topic, payload data, and processing instructions.
// Event is immutable - all modifier methods return a new copy.
type event struct {
	topic    *Topic
	payload  any
	onFinish []func(ctx context.Context)
	wait     bool
	sync     bool
}

// hasOnFinish indicates whether the event has any finish callbacks registered.
// Used internally by the hub to determine if cleanup is needed.
func (e *event) hasOnFinish() bool {
	return len(e.onFinish) > 0
}

// finish executes all registered finish callbacks in sequence.
// Called automatically by the hub after event processing completes.
func (e *event) finish(ctx context.Context) {
	for _, cb := range e.onFinish {
		if cb != nil {
			cb(ctx)
		}
	}
}
