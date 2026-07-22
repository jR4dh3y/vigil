package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/nvr/nvr/server/internal/auth"
)

// wsEnvelope is the JSON frame pushed over the event WebSocket.
type wsEnvelope struct {
	Type string `json:"type"`
	Data Event  `json:"data"`
}

// HandleWS upgrades to a WebSocket and streams event-bus publications as JSON.
// Mounted outside OpenAPI at GET /api/v1/ws.
func (s *Server) HandleWS(w http.ResponseWriter, r *http.Request) {
	u := auth.UserFromContext(r.Context())
	if u == nil {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "unauthorized")
		return
	}
	if s.Event == nil {
		writeError(w, http.StatusInternalServerError, "event service unavailable", "internal")
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		// Dashboard Vite dev origin; production is same-origin.
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		slog.Warn("websocket accept", "err", err)
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	ch, unsub := s.Event.Bus().Subscribe()
	defer unsub()

	ctx := r.Context()
	// Reader loop drains client frames / detects disconnect.
	readDone := make(chan struct{})
	go func() {
		defer close(readDone)
		for {
			_, _, err := conn.Read(ctx)
			if err != nil {
				return
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-readDone:
			return
		case ev, ok := <-ch:
			if !ok {
				return
			}
			payload, err := json.Marshal(wsEnvelope{Type: "event", Data: mapEvent(ev)})
			if err != nil {
				continue
			}
			writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err = conn.Write(writeCtx, websocket.MessageText, payload)
			cancel()
			if err != nil {
				return
			}
		}
	}
}
