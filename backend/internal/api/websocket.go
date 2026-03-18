package api

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hemu1808/auradeploy/backend/internal/models"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

type WsMessage struct {
	Type    string               `json:"type"`
	Payload []models.Application `json:"payload"`
}

func (h *Handler) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("WebSocket Upgrade Error", "error", err)
		return
	}

	clientID := r.RemoteAddr
	slog.Info("New WebSocket Connection", "client", clientID)

	updateCh := make(chan []models.Application, 5)
	h.orch.Subscribe(updateCh)

	// Setup Ping/Pong handling for graceful disconnects
	conn.SetReadLimit(512)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	var wg sync.WaitGroup
	wg.Add(2)

	// Pump 1: Write messages to the client
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(pingPeriod)
		defer func() {
			ticker.Stop()
			conn.Close()
		}()

		for {
			select {
			case apps, ok := <-updateCh:
				conn.SetWriteDeadline(time.Now().Add(writeWait))
				if !ok {
					// The hub closed the channel.
					conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bye"))
					return
				}

				msg := WsMessage{
					Type:    "SYNC_APPS",
					Payload: apps,
				}

				if err := conn.WriteJSON(msg); err != nil {
					slog.Warn("WebSocket write JSON failure", "client", clientID, "error", err)
					return
				}

			case <-ticker.C:
				conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					slog.Info("WebSocket Ping failure (client likely disconnected)", "client", clientID)
					return
				}
			}
		}
	}()

	// Pump 2: Read messages from the client to enforce Pong deadlines
	go func() {
		defer wg.Done()
		defer func() {
			h.orch.Unsubscribe(updateCh)
			conn.Close()
			slog.Info("WebSocket Connection Closed Gracefully", "client", clientID)
		}()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
					slog.Debug("WebSocket closed by client", "client", clientID, "error", err)
				} else {
					slog.Error("WebSocket unexpected close error", "error", err)
				}
				break
			}
		}
	}()

	wg.Wait()
}
