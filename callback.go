package hub

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/spf13/cast"
)

// WrapSubscribeCallback converts various callback signatures into a standardized Event handler function.
// It provides optimized type conversion for supported payload types while maintaining strict type safety.
//
// Supported callback formats:
//  1. Minimal:      func(ctx context.Context) error
//  2. Event style:  func(ctx context.Context, e *Event) error
//  3. Typed payload: func(ctx context.Context, payload Type) error
//  4. Generic payload: func(ctx context.Context, payload any) error
//
// Supported payload types (Type):
//   - All integer types (int, int8, int16, int32, int64)
//   - All unsigned integer types (uint, uint8, uint16, uint32, uint64)
//   - Floating point (float32, float64)
//   - String and boolean (string, bool)
//   - Time and duration (time.Time, time.Duration)
//   - Common collections ([]string, map[string]interface{})
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - cb: The callback function to wrap (must match one of supported signatures)
//
// Returns:
//   - A normalized function with signature func(context.Context, *Event) error
//   - An error if:
//   - Callback is not a function
//   - Invalid return type (must return exactly error)
//   - Invalid parameter count (must be 1-2 parameters)
//   - First parameter is not context.Context
//   - Unsupported parameter type (with details)
//
// Conversion behavior:
//   - For supported types, attempts direct type assertion first
//   - Falls back to github.com/spf13/cast conversion if needed
//   - Ignore conversion errors from cast package
//   - For 'any' type, passes payload through without conversion
//
// Example usage:
//
//	// Minimal callback
//	proxy, err := WrapSubscribeCallback(ctx, func(ctx context.Context) error {
//	    return nil
//	})
//
//	// Event handler
//	proxy, err := WrapSubscribeCallback(ctx, func(ctx context.Context, e *Event) error {
//	    fmt.Printf("Event on %s: %v", e.Topic(), e.Payload())
//	    return nil
//	})
//
//	// Typed payload
//	proxy, err := WrapSubscribeCallback(ctx, func(ctx context.Context, id int) error {
//	    fmt.Printf("Processing ID: %d", id)
//	    return nil
//	})
//
//	// Generic payload
//	proxy, err := WrapSubscribeCallback(ctx, func(ctx context.Context, data any) error {
//	    // Handle any payload type
//	    return nil
//	})
//
// Notes:
// - The function performs all type validation during wrapping, not during event processing
// - For maximum performance with known types, use the specific typed signatures
// - The generic 'any' signature provides flexibility at a small performance cost
func WrapSubscribeCallback(ctx context.Context, cb interface{}) (func(ctx context.Context, e *Event) error, error) {
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
			cbFunc := cb.(func(ctx context.Context, e *Event) error)
			proxy = func(ctx context.Context, e *Event) error {
				return cbFunc(ctx, e)
			}

		// Numeric types
		case reflect.TypeOf(int(0)):
			cbFunc := cb.(func(ctx context.Context, i int) error)
			proxy = func(ctx context.Context, e *Event) error {
				if v, ok := e.Payload().(int); ok {
					return cbFunc(ctx, v)
				}
				return cbFunc(ctx, cast.ToInt(e.Payload()))
			}
		case reflect.TypeOf(int8(0)):
			cbFunc := cb.(func(ctx context.Context, i int8) error)
			proxy = func(ctx context.Context, e *Event) error {
				if v, ok := e.Payload().(int8); ok {
					return cbFunc(ctx, v)
				}
				return cbFunc(ctx, cast.ToInt8(e.Payload()))
			}
		case reflect.TypeOf(int16(0)):
			cbFunc := cb.(func(ctx context.Context, i int16) error)
			proxy = func(ctx context.Context, e *Event) error {
				if v, ok := e.Payload().(int16); ok {
					return cbFunc(ctx, v)
				}
				return cbFunc(ctx, cast.ToInt16(e.Payload()))
			}
		case reflect.TypeOf(int32(0)):
			cbFunc := cb.(func(ctx context.Context, i int32) error)
			proxy = func(ctx context.Context, e *Event) error {
				if v, ok := e.Payload().(int32); ok {
					return cbFunc(ctx, v)
				}
				return cbFunc(ctx, cast.ToInt32(e.Payload()))
			}
		case reflect.TypeOf(int64(0)):
			cbFunc := cb.(func(ctx context.Context, i int64) error)
			proxy = func(ctx context.Context, e *Event) error {
				if v, ok := e.Payload().(int64); ok {
					return cbFunc(ctx, v)
				}
				return cbFunc(ctx, cast.ToInt64(e.Payload()))
			}

		// Unsigned integers
		case reflect.TypeOf(uint(0)):
			cbFunc := cb.(func(ctx context.Context, i uint) error)
			proxy = func(ctx context.Context, e *Event) error {
				if v, ok := e.Payload().(uint); ok {
					return cbFunc(ctx, v)
				}
				return cbFunc(ctx, cast.ToUint(e.Payload()))
			}
		case reflect.TypeOf(uint8(0)):
			cbFunc := cb.(func(ctx context.Context, i uint8) error)
			proxy = func(ctx context.Context, e *Event) error {
				if v, ok := e.Payload().(uint8); ok {
					return cbFunc(ctx, v)
				}
				return cbFunc(ctx, cast.ToUint8(e.Payload()))
			}
		case reflect.TypeOf(uint16(0)):
			cbFunc := cb.(func(ctx context.Context, i uint16) error)
			proxy = func(ctx context.Context, e *Event) error {
				if v, ok := e.Payload().(uint16); ok {
					return cbFunc(ctx, v)
				}
				return cbFunc(ctx, cast.ToUint16(e.Payload()))
			}
		case reflect.TypeOf(uint32(0)):
			cbFunc := cb.(func(ctx context.Context, i uint32) error)
			proxy = func(ctx context.Context, e *Event) error {
				if v, ok := e.Payload().(uint32); ok {
					return cbFunc(ctx, v)
				}
				return cbFunc(ctx, cast.ToUint32(e.Payload()))
			}
		case reflect.TypeOf(uint64(0)):
			cbFunc := cb.(func(ctx context.Context, i uint64) error)
			proxy = func(ctx context.Context, e *Event) error {
				if v, ok := e.Payload().(uint64); ok {
					return cbFunc(ctx, v)
				}
				return cbFunc(ctx, cast.ToUint64(e.Payload()))
			}

		// Floating point
		case reflect.TypeOf(float32(0)):
			cbFunc := cb.(func(ctx context.Context, f float32) error)
			proxy = func(ctx context.Context, e *Event) error {
				if v, ok := e.Payload().(float32); ok {
					return cbFunc(ctx, v)
				}
				return cbFunc(ctx, cast.ToFloat32(e.Payload()))
			}
		case reflect.TypeOf(float64(0)):
			cbFunc := cb.(func(ctx context.Context, f float64) error)
			proxy = func(ctx context.Context, e *Event) error {
				if v, ok := e.Payload().(float64); ok {
					return cbFunc(ctx, v)
				}
				return cbFunc(ctx, cast.ToFloat64(e.Payload()))
			}

		// String and bool
		case reflect.TypeOf(string("")):
			cbFunc := cb.(func(ctx context.Context, s string) error)
			proxy = func(ctx context.Context, e *Event) error {
				if v, ok := e.Payload().(string); ok {
					return cbFunc(ctx, v)
				}
				return cbFunc(ctx, cast.ToString(e.Payload()))
			}
		case reflect.TypeOf(bool(false)):
			cbFunc := cb.(func(ctx context.Context, b bool) error)
			proxy = func(ctx context.Context, e *Event) error {
				if v, ok := e.Payload().(bool); ok {
					return cbFunc(ctx, v)
				}
				return cbFunc(ctx, cast.ToBool(e.Payload()))
			}

		// Time and duration
		case reflect.TypeOf(time.Time{}):
			cbFunc := cb.(func(ctx context.Context, t time.Time) error)
			proxy = func(ctx context.Context, e *Event) error {
				if v, ok := e.Payload().(time.Time); ok {
					return cbFunc(ctx, v)
				}
				return cbFunc(ctx, cast.ToTime(e.Payload()))
			}
		case reflect.TypeOf(time.Duration(0)):
			cbFunc := cb.(func(ctx context.Context, d time.Duration) error)
			proxy = func(ctx context.Context, e *Event) error {
				if v, ok := e.Payload().(time.Duration); ok {
					return cbFunc(ctx, v)
				}
				return cbFunc(ctx, cast.ToDuration(e.Payload()))
			}

		// Slices and maps
		case reflect.TypeOf([]string{}):
			cbFunc := cb.(func(ctx context.Context, s []string) error)
			proxy = func(ctx context.Context, e *Event) error {
				if v, ok := e.Payload().([]string); ok {
					return cbFunc(ctx, v)
				}
				return cbFunc(ctx, cast.ToStringSlice(e.Payload()))
			}
		case reflect.TypeOf(map[string]interface{}{}):
			cbFunc := cb.(func(ctx context.Context, m map[string]interface{}) error)
			proxy = func(ctx context.Context, e *Event) error {
				if v, ok := e.Payload().(map[string]interface{}); ok {
					return cbFunc(ctx, v)
				}
				return cbFunc(ctx, cast.ToStringMap(e.Payload()))
			}

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
