package hub

import (
	"context"
	"slices"
)

// Event represents a message sent to a specific topic in the event hub.
// It contains the topic, payload data, and processing instructions.
// Event is immutable - all modifier methods return a new copy.
type Event struct {
	topic    *Topic
	payload  any
	onFinish []func(ctx context.Context, ev *Event)
	wait     bool
	sync     bool
}

// Topic returns the event's destination topic.
// The topic defines the event's routing path and contains metadata attributes.
func (e *Event) Topic() *Topic {
	return e.topic
}

// Payload returns the event's data content.
// The payload can be of any type and contains the actual message data.
func (e *Event) Payload() any {
	return e.payload
}

// hasOnFinish indicates whether the event has any finish callbacks registered.
// Used internally by the hub to determine if cleanup is needed.
func (e *Event) hasOnFinish() bool {
	return len(e.onFinish) > 0
}

// finish executes all registered finish callbacks in sequence.
// Called automatically by the hub after event processing completes.
func (e *Event) finish(ctx context.Context) {
	for _, cb := range e.onFinish {
		if cb != nil {
			cb(ctx, e)
		}
	}
}

// clone creates a deep copy of the Event.
// Used internally to implement immutability when modifying events.
func (e *Event) clone() *Event {
	n := &Event{
		topic:    e.topic,
		payload:  e.payload,
		onFinish: slices.Clone(e.onFinish),
		wait:     e.wait,
		sync:     e.sync,
	}
	return n
}

// WithOnFinish registers a callback that executes after event processing completes.
// Multiple callbacks can be registered and will execute in registration order.
// The context will reflect the event's final processing state.
//
// Example:
//
//	event.WithOnFinish(func(ctx context.Context, e *Event) {
//	    if ctx.Err() != nil {
//	        log.Println("Event processing failed:", ctx.Err())
//	    }
//	})
func (e *Event) WithOnFinish(cb func(ctx context.Context, ev *Event)) *Event {
	n := e.clone()
	n.onFinish = append(n.onFinish, cb)
	return n
}

// WithPayload sets the event's payload data.
// The payload can be any Go value that needs to be passed with the event.
//
// Example:
//
//	event.WithPayload(map[string]any{"data": 42})
func (e *Event) WithPayload(payload any) *Event {
	n := e.clone()
	n.payload = payload
	return n
}

// WithTopic changes the event's destination topic.
// Useful for redirecting events or adding additional metadata.
//
// Example:
//
//	event.WithTopic(T("type=update", "priority=high"))
func (e *Event) WithTopic(topic *Topic) *Event {
	n := e.clone()
	n.topic = topic
	return n
}

// WithWait sets whether the sender should wait for event processing to complete.
// When true, the Publish method will block until all handlers finish.
// Default is false for asynchronous operation.
func (e *Event) WithWait(v bool) *Event {
	n := e.clone()
	n.wait = v
	return n
}

// WithSync sets whether the event requires synchronous processing.
// Synchronous events are processed sequentially by handlers without goroutines.
// Default is false for concurrent processing.
func (e *Event) WithSync(v bool) *Event {
	n := e.clone()
	n.sync = v
	return n
}

// NewEvent creates a new Event with the specified topic attributes.
// The topic is created from key-value pairs (see NewTopic for format details).
// Returns error if topic attributes are invalid.
//
// Example:
//
//	ev, err := NewEvent("type=alert", "priority=high")
func NewEvent(topicArgs ...string) (*Event, error) {
	t, err := NewTopic(topicArgs...)
	if err != nil {
		return nil, err
	}
	return &Event{
		topic: t,
	}, nil
}

// E creates a new Event with the specified topic attributes, panicking on error.
// Simplified constructor for use in tests and initialization.
//
// Example:
//
//	ev := E("type=alert", "priority=high")
func E(topicArgs ...string) *Event {
	return &Event{
		topic: T(topicArgs...),
	}
}
