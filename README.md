# Hub - Event Pub/Sub Library for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/lomik/hub.svg)](https://pkg.go.dev/github.com/lomik/hub)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Hub is a high-performance publish-subscribe event hub for Go applications, featuring:

- **Topic-based routing** with key-value attributes
- **Efficient matching** using multi-level indexes
- **Flexible subscription** options (sync/async, one-time)
- **Thread-safe** implementation

## Installation

```bash
go get github.com/lomik/hub
```

## Core Concepts
- **Event**: Message with payload and topic attributes
- **Topic**: Routing path defined by key-value pairs (e.g., "type=alert", "priority=high")
- **Subscription**: Callback that receives matching events

## Basic Usage
### Initialization
```go
package main

import "github.com/lomik/hub"

func main() {
    h := hub.New()
}
```

### Publishing Events
```go
// Simple event
h.PublishEvent(ctx, hub.E("type=alert"))

// With payload
h.PublishEvent(ctx, hub.E("type=alert").WithPayload(map[string]any{
    "message": "Server down!",
    "code":    500,
}))

// With options
h.PublishEvent(ctx, hub.E("type=metrics"), 
    hub.Wait(true),       // Wait for all handlers
    hub.Sync(true),       // Process synchronously
    hub.OnFinish(func(ctx context.Context, e *hub.Event) {
        // Cleanup logic
    }),
)
```

### Subscribing to Events
```go
// Subscribe to all events
subID := h.SubscribeEvent(ctx, hub.T(), func(ctx context.Context, e *hub.Event) error {
    fmt.Printf("Received event: %+v\n", e)
    return nil
})

// Subscribe to specific topic
h.SubscribeEvent(ctx, hub.T("type=alert"), func(ctx context.Context, e *hub.Event) error {
    alert := e.Payload().(map[string]any)
    fmt.Println("ALERT:", alert["message"])
    return nil
})

// One-time subscription
h.SubscribeEvent(ctx, hub.T("type=metrics"), func(ctx context.Context, e *hub.Event) error {
    recordMetrics(e.Payload())
    return nil
}, hub.Once(true))
```

### Advanced Patterns

#### Topic Matching
```go
// Subscriber wants alerts of any priority
subscriberTopic := hub.T("type=alert", "priority=*")

// Publisher sends high priority alert
eventTopic := hub.T("type=alert", "priority=high")

// This will match
h.PublishEvent(ctx, &hub.Event{
    topic:   eventTopic,
    payload: "System overload",
})
```

#### Merging Topics
```go
base := hub.T("app=web", "env=production")
extended := base.With("user=admin", "action=delete")

h.PublishEvent(ctx, &hub.Event{
    topic:   extended,
    payload: "User deleted item",
})
```