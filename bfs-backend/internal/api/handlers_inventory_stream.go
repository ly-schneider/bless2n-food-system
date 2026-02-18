package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

func (h *Handlers) StreamInventory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	rc := http.NewResponseController(w)

	ctx := r.Context()
	products, err := h.products.GetAll(ctx)
	if err != nil {
		_, _ = w.Write([]byte("event: error\ndata: {\"error\":\"failed to load products\"}\n\n"))
		_ = rc.Flush()
		return
	}

	ids := make([]uuid.UUID, len(products))
	for i, p := range products {
		ids[i] = p.ID
	}

	stocks, err := h.products.GetStockBatch(ctx, ids)
	if err != nil {
		_, _ = w.Write([]byte("event: error\ndata: {\"error\":\"failed to load stock\"}\n\n"))
		_ = rc.Flush()
		return
	}

	snapshot := make(map[string]int, len(stocks))
	for id, stock := range stocks {
		snapshot[id.String()] = stock
	}

	data, _ := json.Marshal(snapshot)
	_, _ = w.Write(append(append([]byte("event: inventory-snapshot\ndata: "), data...), '\n', '\n'))
	_ = rc.Flush()

	subID := uuid.New().String()
	ch := h.inventoryHub.Subscribe(subID)
	defer h.inventoryHub.Unsubscribe(subID)

	for {
		select {
		case update, ok := <-ch:
			if !ok {
				return
			}
			data, _ := json.Marshal(update)
			_, _ = w.Write(append(append([]byte("event: inventory-update\ndata: "), data...), '\n', '\n'))
			_ = rc.Flush()
		case <-ctx.Done():
			return
		}
	}
}
