package hub

import "context"

type HubOption interface {
	modifyHub(ctx context.Context, h *Hub)
}
