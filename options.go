package hub

import "context"

// SubscribeOption defines an interface for modifying subscription parameters
type SubscribeOption interface {
	modifySub(ctx context.Context, s *sub)
}

// PublishOption defines an interface for modifying event publishing behavior
type PublishOption interface {
	modifyEvent(ctx context.Context, e *Event) *Event
}

// optionSubscribeOnce implements subscription option for single-time delivery
type optionSubscribeOnce struct {
	v bool // Flag indicating whether to deliver only once
}

// modifySub applies the once flag to the subscription
func (o *optionSubscribeOnce) modifySub(ctx context.Context, s *sub) {
	s.once = o.v
}

// Once creates a SubscribeOption that controls single delivery
// When true, the subscription will be automatically removed after first delivery
func Once(v bool) SubscribeOption {
	return &optionSubscribeOnce{
		v: v,
	}
}

// optionPublishSync implements synchronous publishing option
type optionPublishSync struct {
	v bool // Flag indicating synchronous processing
}

// modifyEvent applies synchronous processing flag to the event
func (o *optionPublishSync) modifyEvent(ctx context.Context, e *Event) *Event {
	return e.WithSync(o.v)
}

// Sync creates a PublishOption that controls synchronous event processing.
// When enabled (v=true), the event will be processed by all handlers sequentially
// in the publishing goroutine, without spawning additional goroutines.
// This ensures strict ordering but blocks the publisher during processing.
func Sync(v bool) PublishOption {
	return &optionPublishSync{
		v: v,
	}
}

// optionPublishWait implements waiting option for publish completion
type optionPublishWait struct {
	v bool // Flag indicating whether to wait for completion
}

// modifyEvent applies wait flag to the event
func (o *optionPublishWait) modifyEvent(ctx context.Context, e *Event) *Event {
	return e.WithWait(o.v)
}

// Wait creates a PublishOption that controls waiting for handlers
// When true, PublishEvent will block until all handlers complete
func Wait(v bool) PublishOption {
	return &optionPublishWait{
		v: v,
	}
}

// optionPublishOnFinish implements callback after publish completion
type optionPublishOnFinish struct {
	cb func(ctx context.Context, ev *Event) // Callback function
}

// modifyEvent adds completion callback to the event
func (o *optionPublishOnFinish) modifyEvent(ctx context.Context, e *Event) *Event {
	if o.cb == nil {
		return e
	}
	return e.WithOnFinish(o.cb)
}

// OnFinish creates a PublishOption with completion callback
// The callback executes after all handlers process the event
func OnFinish(cb func(ctx context.Context, ev *Event)) PublishOption {
	return &optionPublishOnFinish{
		cb: cb,
	}
}
