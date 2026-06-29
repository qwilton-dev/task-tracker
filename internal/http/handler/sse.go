package handler

import (
	"fmt"
	"net/http"
	"task-tracker/internal/events"

	"github.com/go-chi/chi/v5"
)

type SSEHandler struct {
	Hub *events.Hub
}

func NewSSEHandler(hub *events.Hub) *SSEHandler {
	return &SSEHandler{Hub: hub}
}

func (h *SSEHandler) Stream(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "no_streaming_support", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	flusher.Flush()

	eventCh := h.Hub.Subscribe(projectID)
	defer h.Hub.Unsubscribe(projectID, eventCh)

	ctx := r.Context()
	for {
		select {
		case event := <-eventCh:
			_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, string(event.Payload))
			flusher.Flush()
		case <-ctx.Done():
			return
		}
	}
}
