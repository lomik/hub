# Hub - Event Pub/Sub Library for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/lomik/hub.svg)](https://pkg.go.dev/github.com/lomik/hub)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Hub is a publish-subscribe event hub for Go applications, featuring:

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

The `Publish` method provides a streamlined way to send events:

```go
// Simple event with string payload
h.Publish(ctx, 
    hub.T("type=alert", "priority=high"),
    "Server CPU overload",
)

// With structured data and options
h.Publish(ctx,
    hub.T("type=metrics", "source=api"),
    map[string]any{
        "cpu": 85.2,
        "mem": 45.7,
    },
    hub.Wait(true), // Wait for handlers
    hub.Sync(true), // Process synchronously
)
```

### Subscribing to Events
The `Subscribe` method supports flexible callback signatures:

```go
// Minimal callback
id1, _ := h.Subscribe(ctx, hub.T("type=alert"), func(ctx context.Context) error {
    log.Println("Alert received")
    return nil
})

// Typed payload
id2, _ := h.Subscribe(ctx, hub.T("type=metrics"), func(ctx context.Context, stats map[string]float64) error {
    log.Printf("CPU: %.1f%%, Mem: %.1f%%", stats["cpu"], stats["mem"])
    return nil
})

// Generic payload
id3, _ := h.Subscribe(ctx, hub.T(), func(ctx context.Context, payload any) error {
    log.Printf("Event received: %T", payload)
    return nil
})
```

### Complete Example
```go
h := hub.New()

// Subscribe to metrics
h.Subscribe(ctx, hub.T("type=metrics"), func(ctx context.Context, m map[string]any) error {
    log.Printf("Metrics: %+v", m)
    return nil
})

// Publish metrics
h.Publish(ctx,
    hub.T("type=metrics"),
    map[string]any{
        "requests": 1423,
        "errors":   27,
    },
)

// Unsubscribe later
h.Unsubscribe(ctx, id)
```

### Patterns

#### Topic Matching
```go
// Subscriber wants alerts of any priority
subscriberTopic := hub.T("type=alert", "priority=*")

// Publisher sends high priority alert
eventTopic := hub.T("type=alert", "priority=high")

// This will match
h.Subscribe(ctx, subscriberTopic, func(ctx context.Context, p string) error {
    fmt.Println(p)
})
h.Publish(ctx, eventTopic, "System overload")
```

#### Merging Topics
```go
base := hub.T("app=web", "env=production")
extended := base.With("user=admin", "action=delete")

h.Publish(ctx, extended, "User deleted item")
```
