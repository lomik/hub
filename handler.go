package hub

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cast"
)

// Handler defines a function signature for processing events in the hub.
// It receives a context for cancellation/timeout and the event to process.
// Return an error to indicate processing failure.
//
// Usage:
//
//	var myHandler Handler = func(ctx context.Context, e *Event) error {
//	    log.Printf("Processing event: %s", e.Topic())
//	    return nil // Return nil on success
//	}
//
//	h.Subscribe(ctx, topic, myHandler)
type Handler func(ctx context.Context, t *Topic, p any) error

func toHandlerCommon[T any](cb func(context.Context, T) error, castFunc func(any) T) Handler {
	return func(ctx context.Context, t *Topic, p any) error {
		if v, ok := p.(T); ok {
			return cb(ctx, v)
		}
		return cb(ctx, castFunc(p))
	}
}

// ToHandler converts various callback signatures into a standardized Event handler function.
func (h *Hub) ToHandler(ctx context.Context, cb any) (Handler, error) {
	// custom converters
	for _, c := range h.convertToHandler {
		ret, err := c(ctx, cb)
		if err != nil {
			return nil, err
		}
		if ret != nil {
			return ret, nil
		}
	}

	switch cbt := cb.(type) {
	case func(ctx context.Context) error:
		return func(ctx context.Context, t *Topic, p any) error {
			return cbt(ctx)
		}, nil

	case Handler:
		return cbt, nil
	case func(context.Context, *Topic, any) error:
		return cbt, nil

	// Numeric types
	case func(context.Context, int) error:
		return toHandlerCommon(cbt, cast.ToInt), nil
	case func(context.Context, int8) error:
		return toHandlerCommon(cbt, cast.ToInt8), nil
	case func(context.Context, int16) error:
		return toHandlerCommon(cbt, cast.ToInt16), nil
	case func(context.Context, int32) error:
		return toHandlerCommon(cbt, cast.ToInt32), nil
	case func(context.Context, int64) error:
		return toHandlerCommon(cbt, cast.ToInt64), nil

	// Unsigned integers
	case func(context.Context, uint) error:
		return toHandlerCommon(cbt, cast.ToUint), nil
	case func(context.Context, uint8) error:
		return toHandlerCommon(cbt, cast.ToUint8), nil
	case func(context.Context, uint16) error:
		return toHandlerCommon(cbt, cast.ToUint16), nil
	case func(context.Context, uint32) error:
		return toHandlerCommon(cbt, cast.ToUint32), nil
	case func(context.Context, uint64) error:
		return toHandlerCommon(cbt, cast.ToUint64), nil

	// Floating point
	case func(context.Context, float32) error:
		return toHandlerCommon(cbt, cast.ToFloat32), nil
	case func(context.Context, float64) error:
		return toHandlerCommon(cbt, cast.ToFloat64), nil

	// String and bool
	case func(context.Context, string) error:
		return toHandlerCommon(cbt, cast.ToString), nil
	case func(context.Context, bool) error:
		return toHandlerCommon(cbt, cast.ToBool), nil

	// Time and duration
	case func(context.Context, time.Time) error:
		return toHandlerCommon(cbt, cast.ToTime), nil
	case func(context.Context, time.Duration) error:
		return toHandlerCommon(cbt, cast.ToDuration), nil

	// Slices and maps
	case func(context.Context, []string) error:
		return toHandlerCommon(cbt, cast.ToStringSlice), nil
	case func(context.Context, map[string]any) error:
		return toHandlerCommon(cbt, cast.ToStringMap), nil
	case func(ctx context.Context, a any) error:
		return func(ctx context.Context, t *Topic, p any) error {
			return cbt(ctx, p)
		}, nil

	// default
	default:
		// Return error for unsupported types
		return nil, fmt.Errorf("unsupported callback type: %T", cb)
	}
}
