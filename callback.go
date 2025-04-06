package hub

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/spf13/cast"
)

func wrapSubscribeCallbackGeneric[T any](cb any, castFunc func(any) T) func(ctx context.Context, e *Event) error {
	cbFunc := cb.(func(context.Context, T) error)
	return func(ctx context.Context, e *Event) error {
		if v, ok := e.Payload().(T); ok {
			return cbFunc(ctx, v)
		}
		return cbFunc(ctx, castFunc(e.Payload()))
	}
}

// wrapSubscribeCallback converts various callback signatures into a standardized Event handler function.
func wrapSubscribeCallback(ctx context.Context, cb any) (func(ctx context.Context, e *Event) error, error) {
	cbVal := reflect.ValueOf(cb)
	if cbVal.Kind() != reflect.Func {
		return nil, fmt.Errorf("callback must be a function")
	}

	cbType := cbVal.Type()

	// Validate return type
	if cbType.NumOut() != 1 || !cbType.Out(0).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return nil, fmt.Errorf("callback must return exactly one error value")
	}

	// Validate input parameters
	numIn := cbType.NumIn()
	if numIn < 1 || numIn > 2 {
		return nil, fmt.Errorf("callback must have 1-2 parameters")
	}

	// First parameter must be context.Context
	if cbType.In(0) != reflect.TypeOf((*context.Context)(nil)).Elem() {
		return nil, fmt.Errorf("first parameter must be context.Context")
	}

	var proxy func(context.Context, *Event) error

	switch {
	case numIn == 1:
		// Format: func(ctx context.Context) error
		cbFunc := cb.(func(ctx context.Context) error)
		proxy = func(ctx context.Context, e *Event) error {
			return cbFunc(ctx)
		}

	case numIn == 2:
		// Format:
		// func(ctx context.Context, e *Event) error
		// func(ctx context.Context, payload Type) error

		switch paramType := cbType.In(1); paramType {
		case reflect.TypeOf((*Event)(nil)):
			proxy = cb.(func(ctx context.Context, e *Event) error)

		// Numeric types
		case reflect.TypeOf(int(0)):
			proxy = wrapSubscribeCallbackGeneric(cb, cast.ToInt)
		case reflect.TypeOf(int8(0)):
			proxy = wrapSubscribeCallbackGeneric(cb, cast.ToInt8)
		case reflect.TypeOf(int16(0)):
			proxy = wrapSubscribeCallbackGeneric(cb, cast.ToInt16)
		case reflect.TypeOf(int32(0)):
			proxy = wrapSubscribeCallbackGeneric(cb, cast.ToInt32)
		case reflect.TypeOf(int64(0)):
			proxy = wrapSubscribeCallbackGeneric(cb, cast.ToInt64)

		// Unsigned integers
		case reflect.TypeOf(uint(0)):
			proxy = wrapSubscribeCallbackGeneric(cb, cast.ToUint)
		case reflect.TypeOf(uint8(0)):
			proxy = wrapSubscribeCallbackGeneric(cb, cast.ToUint8)
		case reflect.TypeOf(uint16(0)):
			proxy = wrapSubscribeCallbackGeneric(cb, cast.ToUint16)
		case reflect.TypeOf(uint32(0)):
			proxy = wrapSubscribeCallbackGeneric(cb, cast.ToUint32)
		case reflect.TypeOf(uint64(0)):
			proxy = wrapSubscribeCallbackGeneric(cb, cast.ToUint64)

		// Floating point
		case reflect.TypeOf(float32(0)):
			proxy = wrapSubscribeCallbackGeneric(cb, cast.ToFloat32)
		case reflect.TypeOf(float64(0)):
			proxy = wrapSubscribeCallbackGeneric(cb, cast.ToFloat64)

		// String and bool
		case reflect.TypeOf(string("")):
			proxy = wrapSubscribeCallbackGeneric(cb, cast.ToString)
		case reflect.TypeOf(bool(false)):
			proxy = wrapSubscribeCallbackGeneric(cb, cast.ToBool)

		// Time and duration
		case reflect.TypeOf(time.Time{}):
			proxy = wrapSubscribeCallbackGeneric(cb, cast.ToTime)
		case reflect.TypeOf(time.Duration(0)):
			proxy = wrapSubscribeCallbackGeneric(cb, cast.ToDuration)

		// Slices and maps
		case reflect.TypeOf([]string{}):
			proxy = wrapSubscribeCallbackGeneric(cb, cast.ToStringSlice)
		case reflect.TypeOf(map[string]any{}):
			proxy = wrapSubscribeCallbackGeneric(cb, cast.ToStringMap)

		default:
			if cbFunc, ok := cb.(func(ctx context.Context, a any) error); ok {
				proxy = func(ctx context.Context, e *Event) error {
					return cbFunc(ctx, e.Payload())
				}
			} else {
				// Return error for unsupported types
				return nil, fmt.Errorf("unsupported parameter type: %v", paramType)
			}
		}
	}

	return proxy, nil
}
