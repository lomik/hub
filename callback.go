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

// wrapSubscribeCallback converts various callback signatures into a standardized Event handler function.
func wrapSubscribeCallback(_ context.Context, cb any) (func(ctx context.Context, e *Event) error, error) {
	var proxy func(context.Context, *Event) error

	switch cbt := cb.(type) {
	case func(ctx context.Context) error:
		proxy = func(ctx context.Context, e *Event) error {
			return cbt(ctx)
		}

	case func(context.Context, *Event) error:
		proxy = cbt

	// Numeric types
	case func(context.Context, int) error:
		proxy = wrapSubscribeCallbackGeneric(cbt, cast.ToInt)
	case func(context.Context, int8) error:
		proxy = wrapSubscribeCallbackGeneric(cbt, cast.ToInt8)
	case func(context.Context, int16) error:
		proxy = wrapSubscribeCallbackGeneric(cbt, cast.ToInt16)
	case func(context.Context, int32) error:
		proxy = wrapSubscribeCallbackGeneric(cbt, cast.ToInt32)
	case func(context.Context, int64) error:
		proxy = wrapSubscribeCallbackGeneric(cbt, cast.ToInt64)

	// Unsigned integers
	case func(context.Context, uint) error:
		proxy = wrapSubscribeCallbackGeneric(cbt, cast.ToUint)
	case func(context.Context, uint8) error:
		proxy = wrapSubscribeCallbackGeneric(cbt, cast.ToUint8)
	case func(context.Context, uint16) error:
		proxy = wrapSubscribeCallbackGeneric(cbt, cast.ToUint16)
	case func(context.Context, uint32) error:
		proxy = wrapSubscribeCallbackGeneric(cbt, cast.ToUint32)
	case func(context.Context, uint64) error:
		proxy = wrapSubscribeCallbackGeneric(cbt, cast.ToUint64)

	// Floating point
	case func(context.Context, float32) error:
		proxy = wrapSubscribeCallbackGeneric(cbt, cast.ToFloat32)
	case func(context.Context, float64) error:
		proxy = wrapSubscribeCallbackGeneric(cbt, cast.ToFloat64)

	// String and bool
	case func(context.Context, string) error:
		proxy = wrapSubscribeCallbackGeneric(cbt, cast.ToString)
	case func(context.Context, bool) error:
		proxy = wrapSubscribeCallbackGeneric(cbt, cast.ToBool)

	// Time and duration
	case func(context.Context, time.Time) error:
		proxy = wrapSubscribeCallbackGeneric(cbt, cast.ToTime)
	case func(context.Context, time.Duration) error:
		proxy = wrapSubscribeCallbackGeneric(cbt, cast.ToDuration)

	// Slices and maps
	case func(context.Context, []string) error:
		proxy = wrapSubscribeCallbackGeneric(cbt, cast.ToStringSlice)
	case func(context.Context, map[string]any) error:
		proxy = wrapSubscribeCallbackGeneric(cbt, cast.ToStringMap)
	case func(ctx context.Context, a any) error:
		proxy = func(ctx context.Context, e *Event) error {
			return cbt(ctx, e.Payload())
		}
	default:
		// Return error for unsupported types
		return nil, fmt.Errorf("unsupported callback type: %T", cb)
	}

	return proxy, nil
}
