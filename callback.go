package hub

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cast"
)

func wrapSubscribeCallbackGeneric[T any](cb func(context.Context, T) error, castFunc func(any) T) func(ctx context.Context, e *Event) error {
	return func(ctx context.Context, e *Event) error {
		if v, ok := e.Payload().(T); ok {
			return cb(ctx, v)
		}
		return cb(ctx, castFunc(e.Payload()))
	}
}

// WrapSubscribeCallback converts various callback signatures into a standardized Event handler function.
func (h *Hub) WrapSubscribeCallback(_ context.Context, cb any) (func(ctx context.Context, e *Event) error, error) {
	switch cbt := cb.(type) {
	case func(ctx context.Context) error:
		return func(ctx context.Context, e *Event) error {
			return cbt(ctx)
		}, nil

	case func(context.Context, *Event) error:
		return cbt, nil

	// Numeric types
	case func(context.Context, int) error:
		return wrapSubscribeCallbackGeneric(cbt, cast.ToInt), nil
	case func(context.Context, int8) error:
		return wrapSubscribeCallbackGeneric(cbt, cast.ToInt8), nil
	case func(context.Context, int16) error:
		return wrapSubscribeCallbackGeneric(cbt, cast.ToInt16), nil
	case func(context.Context, int32) error:
		return wrapSubscribeCallbackGeneric(cbt, cast.ToInt32), nil
	case func(context.Context, int64) error:
		return wrapSubscribeCallbackGeneric(cbt, cast.ToInt64), nil

	// Unsigned integers
	case func(context.Context, uint) error:
		return wrapSubscribeCallbackGeneric(cbt, cast.ToUint), nil
	case func(context.Context, uint8) error:
		return wrapSubscribeCallbackGeneric(cbt, cast.ToUint8), nil
	case func(context.Context, uint16) error:
		return wrapSubscribeCallbackGeneric(cbt, cast.ToUint16), nil
	case func(context.Context, uint32) error:
		return wrapSubscribeCallbackGeneric(cbt, cast.ToUint32), nil
	case func(context.Context, uint64) error:
		return wrapSubscribeCallbackGeneric(cbt, cast.ToUint64), nil

	// Floating point
	case func(context.Context, float32) error:
		return wrapSubscribeCallbackGeneric(cbt, cast.ToFloat32), nil
	case func(context.Context, float64) error:
		return wrapSubscribeCallbackGeneric(cbt, cast.ToFloat64), nil

	// String and bool
	case func(context.Context, string) error:
		return wrapSubscribeCallbackGeneric(cbt, cast.ToString), nil
	case func(context.Context, bool) error:
		return wrapSubscribeCallbackGeneric(cbt, cast.ToBool), nil

	// Time and duration
	case func(context.Context, time.Time) error:
		return wrapSubscribeCallbackGeneric(cbt, cast.ToTime), nil
	case func(context.Context, time.Duration) error:
		return wrapSubscribeCallbackGeneric(cbt, cast.ToDuration), nil

	// Slices and maps
	case func(context.Context, []string) error:
		return wrapSubscribeCallbackGeneric(cbt, cast.ToStringSlice), nil
	case func(context.Context, map[string]any) error:
		return wrapSubscribeCallbackGeneric(cbt, cast.ToStringMap), nil
	case func(ctx context.Context, a any) error:
		return func(ctx context.Context, e *Event) error {
			return cbt(ctx, e.Payload())
		}, nil

	// default
	default:
		// Return error for unsupported types
		return nil, fmt.Errorf("unsupported callback type: %T", cb)
	}
}
