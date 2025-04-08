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

func toHandlerWithError[T any](cb func(context.Context, T) error, castFunc func(any) T) Handler {
	return func(ctx context.Context, t *Topic, p any) error {
		if v, ok := p.(T); ok {
			return cb(ctx, v)
		}
		return cb(ctx, castFunc(p))
	}
}

func toHandlerNoError[T any](cb func(context.Context, T), castFunc func(any) T) Handler {
	return func(ctx context.Context, t *Topic, p any) error {
		if v, ok := p.(T); ok {
			cb(ctx, v)
			return nil
		}
		cb(ctx, castFunc(p))
		return nil
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
	case func(ctx context.Context):
		return func(ctx context.Context, t *Topic, p any) error {
			cbt(ctx)
			return nil
		}, nil

	case Handler:
		return cbt, nil
	case func(context.Context, *Topic, any) error:
		return cbt, nil
	case func(ctx context.Context, t *Topic, p any):
		return func(ctx context.Context, t *Topic, p any) error {
			cbt(ctx, t, p)
			return nil
		}, nil

	// Numeric types
	case func(context.Context, int) error:
		return toHandlerWithError(cbt, cast.ToInt), nil
	case func(context.Context, int):
		return toHandlerNoError(cbt, cast.ToInt), nil
	case func(context.Context, int8) error:
		return toHandlerWithError(cbt, cast.ToInt8), nil
	case func(context.Context, int8):
		return toHandlerNoError(cbt, cast.ToInt8), nil
	case func(context.Context, int16) error:
		return toHandlerWithError(cbt, cast.ToInt16), nil
	case func(context.Context, int16):
		return toHandlerNoError(cbt, cast.ToInt16), nil
	case func(context.Context, int32) error:
		return toHandlerWithError(cbt, cast.ToInt32), nil
	case func(context.Context, int32):
		return toHandlerNoError(cbt, cast.ToInt32), nil
	case func(context.Context, int64) error:
		return toHandlerWithError(cbt, cast.ToInt64), nil
	case func(context.Context, int64):
		return toHandlerNoError(cbt, cast.ToInt64), nil

	// Unsigned integers
	case func(context.Context, uint) error:
		return toHandlerWithError(cbt, cast.ToUint), nil
	case func(context.Context, uint):
		return toHandlerNoError(cbt, cast.ToUint), nil
	case func(context.Context, uint8) error:
		return toHandlerWithError(cbt, cast.ToUint8), nil
	case func(context.Context, uint8):
		return toHandlerNoError(cbt, cast.ToUint8), nil
	case func(context.Context, uint16) error:
		return toHandlerWithError(cbt, cast.ToUint16), nil
	case func(context.Context, uint16):
		return toHandlerNoError(cbt, cast.ToUint16), nil
	case func(context.Context, uint32) error:
		return toHandlerWithError(cbt, cast.ToUint32), nil
	case func(context.Context, uint32):
		return toHandlerNoError(cbt, cast.ToUint32), nil
	case func(context.Context, uint64) error:
		return toHandlerWithError(cbt, cast.ToUint64), nil
	case func(context.Context, uint64):
		return toHandlerNoError(cbt, cast.ToUint64), nil

	// Floating point
	case func(context.Context, float32) error:
		return toHandlerWithError(cbt, cast.ToFloat32), nil
	case func(context.Context, float32):
		return toHandlerNoError(cbt, cast.ToFloat32), nil
	case func(context.Context, float64) error:
		return toHandlerWithError(cbt, cast.ToFloat64), nil
	case func(context.Context, float64):
		return toHandlerNoError(cbt, cast.ToFloat64), nil

	// String and bool
	case func(context.Context, string) error:
		return toHandlerWithError(cbt, cast.ToString), nil
	case func(context.Context, string):
		return toHandlerNoError(cbt, cast.ToString), nil
	case func(context.Context, bool) error:
		return toHandlerWithError(cbt, cast.ToBool), nil
	case func(context.Context, bool):
		return toHandlerNoError(cbt, cast.ToBool), nil

	// Time and duration
	case func(context.Context, time.Time) error:
		return toHandlerWithError(cbt, cast.ToTime), nil
	case func(context.Context, time.Time):
		return toHandlerNoError(cbt, cast.ToTime), nil
	case func(context.Context, time.Duration) error:
		return toHandlerWithError(cbt, cast.ToDuration), nil
	case func(context.Context, time.Duration):
		return toHandlerNoError(cbt, cast.ToDuration), nil

	// Slices and maps
	case func(context.Context, []string) error:
		return toHandlerWithError(cbt, cast.ToStringSlice), nil
	case func(context.Context, []string):
		return toHandlerNoError(cbt, cast.ToStringSlice), nil
	case func(context.Context, map[string]any) error:
		return toHandlerWithError(cbt, cast.ToStringMap), nil
	case func(context.Context, map[string]any):
		return toHandlerNoError(cbt, cast.ToStringMap), nil
	case func(ctx context.Context, a any) error:
		return func(ctx context.Context, t *Topic, p any) error {
			return cbt(ctx, p)
		}, nil
	case func(ctx context.Context, a any):
		return func(ctx context.Context, t *Topic, p any) error {
			cbt(ctx, p)
			return nil
		}, nil

	// default
	default:
		// Return error for unsupported types
		return nil, fmt.Errorf("unsupported callback type: %T", cb)
	}
}
