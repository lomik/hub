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

func toHandlerWithError[T any](cb func(context.Context, T) error, castFunc func(any) (T, error)) Handler {
	return func(ctx context.Context, t *Topic, p any) error {
		if v, ok := p.(T); ok {
			return cb(ctx, v)
		}
		v, err := castFunc(p)
		if err != nil {
			return newCastError(err)
		}
		return cb(ctx, v)
	}
}

func toHandlerNoError[T any](cb func(context.Context, T), castFunc func(any) (T, error)) Handler {
	return func(ctx context.Context, t *Topic, p any) error {
		if v, ok := p.(T); ok {
			cb(ctx, v)
			return nil
		}
		v, err := castFunc(p)
		if err != nil {
			return newCastError(err)
		}
		cb(ctx, v)
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
		return toHandlerWithError(cbt, cast.ToIntE), nil
	case func(context.Context, int):
		return toHandlerNoError(cbt, cast.ToIntE), nil
	case func(context.Context, int8) error:
		return toHandlerWithError(cbt, cast.ToInt8E), nil
	case func(context.Context, int8):
		return toHandlerNoError(cbt, cast.ToInt8E), nil
	case func(context.Context, int16) error:
		return toHandlerWithError(cbt, cast.ToInt16E), nil
	case func(context.Context, int16):
		return toHandlerNoError(cbt, cast.ToInt16E), nil
	case func(context.Context, int32) error:
		return toHandlerWithError(cbt, cast.ToInt32E), nil
	case func(context.Context, int32):
		return toHandlerNoError(cbt, cast.ToInt32E), nil
	case func(context.Context, int64) error:
		return toHandlerWithError(cbt, cast.ToInt64E), nil
	case func(context.Context, int64):
		return toHandlerNoError(cbt, cast.ToInt64E), nil

	// Unsigned integers
	case func(context.Context, uint) error:
		return toHandlerWithError(cbt, cast.ToUintE), nil
	case func(context.Context, uint):
		return toHandlerNoError(cbt, cast.ToUintE), nil
	case func(context.Context, uint8) error:
		return toHandlerWithError(cbt, cast.ToUint8E), nil
	case func(context.Context, uint8):
		return toHandlerNoError(cbt, cast.ToUint8E), nil
	case func(context.Context, uint16) error:
		return toHandlerWithError(cbt, cast.ToUint16E), nil
	case func(context.Context, uint16):
		return toHandlerNoError(cbt, cast.ToUint16E), nil
	case func(context.Context, uint32) error:
		return toHandlerWithError(cbt, cast.ToUint32E), nil
	case func(context.Context, uint32):
		return toHandlerNoError(cbt, cast.ToUint32E), nil
	case func(context.Context, uint64) error:
		return toHandlerWithError(cbt, cast.ToUint64E), nil
	case func(context.Context, uint64):
		return toHandlerNoError(cbt, cast.ToUint64E), nil

	// Floating point
	case func(context.Context, float32) error:
		return toHandlerWithError(cbt, cast.ToFloat32E), nil
	case func(context.Context, float32):
		return toHandlerNoError(cbt, cast.ToFloat32E), nil
	case func(context.Context, float64) error:
		return toHandlerWithError(cbt, cast.ToFloat64E), nil
	case func(context.Context, float64):
		return toHandlerNoError(cbt, cast.ToFloat64E), nil

	// String and bool
	case func(context.Context, string) error:
		return toHandlerWithError(cbt, cast.ToStringE), nil
	case func(context.Context, string):
		return toHandlerNoError(cbt, cast.ToStringE), nil
	case func(context.Context, bool) error:
		return toHandlerWithError(cbt, cast.ToBoolE), nil
	case func(context.Context, bool):
		return toHandlerNoError(cbt, cast.ToBoolE), nil

	// Time and duration
	case func(context.Context, time.Time) error:
		return toHandlerWithError(cbt, cast.ToTimeE), nil
	case func(context.Context, time.Time):
		return toHandlerNoError(cbt, cast.ToTimeE), nil
	case func(context.Context, time.Duration) error:
		return toHandlerWithError(cbt, cast.ToDurationE), nil
	case func(context.Context, time.Duration):
		return toHandlerNoError(cbt, cast.ToDurationE), nil

	// Slices and maps
	case func(context.Context, []string) error:
		return toHandlerWithError(cbt, cast.ToStringSliceE), nil
	case func(context.Context, []string):
		return toHandlerNoError(cbt, cast.ToStringSliceE), nil
	case func(context.Context, map[string]any) error:
		return toHandlerWithError(cbt, cast.ToStringMapE), nil
	case func(context.Context, map[string]any):
		return toHandlerNoError(cbt, cast.ToStringMapE), nil
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
