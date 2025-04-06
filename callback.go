package hub

import (
	"context"
	"fmt"
	"reflect"
)

// WrapSubscribeCallback converts various callback signatures into a standardized Event handler.
//
// This function accepts different callback styles and returns a normalized function with
// signature func(context.Context, *Event) error that can be used with SubscribeEvent.
//
// Supported callback formats:
//  1. Minimal:      func(ctx context.Context) error
//  2. Payload-only: func(ctx context.Context, payload any) error
//  3. Full context: func(ctx context.Context, topic *Topic, payload any) error
//  4. Event style:  func(ctx context.Context, e *Event) error
//
// Parameters:
//   - ctx: Context for cancellation and timeouts (passed through to callback)
//   - cb:  The callback function to wrap (must match one of supported signatures)
//
// Returns:
//   - A normalized function with signature func(context.Context, *Event) error
//   - An error if the callback signature is invalid
//
// Example usage:
//
//	// As payload handler
//	proxy, err := WrapSubscribeCallback(ctx, func(ctx context.Context, data any) error {
//	    fmt.Println("Received:", data)
//	    return nil
//	})
//
//	// As full event handler
//	proxy, err := WrapSubscribeCallback(ctx, func(ctx context.Context, e *Event) error {
//	    fmt.Printf("Event on %s: %v", e.Topic(), e.Payload())
//	    return nil
//	})
//
// Notes:
// - The returned proxy function maintains the original callback's error handling
// - All parameter type checks are done at wrap time, not during event processing
// - The context parameter is always required as first argument
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
	if numIn < 1 || numIn > 3 {
		return nil, fmt.Errorf("callback must have 1-3 parameters")
	}

	// First parameter must be context.Context
	if cbType.In(0) != reflect.TypeOf((*context.Context)(nil)).Elem() {
		return nil, fmt.Errorf("first parameter must be context.Context")
	}

	var proxy func(context.Context, *Event) error

	switch {
	case numIn == 1:
		// Format: func(ctx context.Context) error
		proxy = func(ctx context.Context, e *Event) error {
			out := cbVal.Call([]reflect.Value{reflect.ValueOf(ctx)})
			if err := out[0].Interface(); err != nil {
				return err.(error)
			}
			return nil
		}

	case numIn == 2:
		// Check second parameter type
		switch cbType.In(1) {
		case reflect.TypeOf((*Event)(nil)):
			// Format: func(ctx context.Context, e *Event) error
			proxy = func(ctx context.Context, e *Event) error {
				out := cbVal.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(e)})
				if err := out[0].Interface(); err != nil {
					return err.(error)
				}
				return nil
			}
		default:
			// Format: func(ctx context.Context, payload any) error
			proxy = func(ctx context.Context, e *Event) error {
				out := cbVal.Call([]reflect.Value{
					reflect.ValueOf(ctx),
					reflect.ValueOf(e.Payload()),
				})
				if err := out[0].Interface(); err != nil {
					return err.(error)
				}
				return nil
			}
		}

	case numIn == 3:
		// Format: func(ctx context.Context, topic *Topic, payload any) error
		if cbType.In(1) != reflect.TypeOf((*Topic)(nil)) {
			return nil, fmt.Errorf("second parameter must be *Topic when using three parameters")
		}

		proxy = func(ctx context.Context, e *Event) error {
			out := cbVal.Call([]reflect.Value{
				reflect.ValueOf(ctx),
				reflect.ValueOf(e.Topic()),
				reflect.ValueOf(e.Payload()),
			})
			if err := out[0].Interface(); err != nil {
				return err.(error)
			}
			return nil
		}
	}

	return proxy, nil
}
