package hub

import "context"

// HubOption defines an interface for configuring Hub instances during creation.
type HubOption interface {
	modifyHub(h *Hub)
}

// ToHandler registers a custom callback converter function that transforms
// arbitrary functions into Handler types. This enables support for additional
// callback signatures beyond the built-in ones.
//
// The converter function should:
//   - Accept a context and the callback value
//   - Return a Handler and nil error if conversion succeeds
//   - Return an error if something wrong
//   - Return nil Handler and nil error if the callback type is unsupported
//
// Multiple converters can be registered. They are tried in registration order
// until one succeeds or all fail.
//
// Example:
//
//	hub.New(
//	    hub.ToHandler(func(ctx context.Context, cb any) (Handler, error) {
//	        if fn, ok := cb.(func(string) error); ok {
//	            return func(ctx context.Context, e *Event) error {
//	                s, _ := e.Payload().(string)
//	                return fn(s)
//	            }, nil
//	        }
//	        return nil, errors.New("unsupported type")
//	    }),
//	)
func ToHandler(converter func(_ context.Context, cb any) (Handler, error)) HubOption {
	return &optionHubToHandler{
		v: converter,
	}
}

// optionHubToHandler implements the HubOption interface for callback converters
type optionHubToHandler struct {
	v func(_ context.Context, cb any) (Handler, error)
}

// modifyHub applies the converter registration to the Hub instance
func (o *optionHubToHandler) modifyHub(h *Hub) {
	if o.v != nil {
		h.convertToHandler = append(h.convertToHandler, o.v)
	}
}
