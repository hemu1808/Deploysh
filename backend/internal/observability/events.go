package observability

import (
	"log/slog"
	"time"
)

// EventType categorizes events (Normal vs Warning)
type EventType string

const (
	EventTypeNormal  EventType = "Normal"
	EventTypeWarning EventType = "Warning"
)

// Event is a Kubernetes-style structured event representing a cluster state transition
type Event struct {
	ID        string    `json:"id"`
	Type      EventType `json:"type"`
	Reason    string    `json:"reason"`
	Message   string    `json:"message"`
	Component string    `json:"component"`
	Timestamp int64     `json:"timestamp"`
}

// EventRecorder provides a centralized way to broadcast and persist events
type EventRecorder struct {
	logger *slog.Logger
}

func NewEventRecorder(logger *slog.Logger) *EventRecorder {
	return &EventRecorder{logger: logger}
}

// RecordEvent logs an event. In a complete production system, this could also write the event into the Raft store.
func (e *EventRecorder) RecordEvent(eventType EventType, reason, message, component string) {
	ev := Event{
		ID:        time.Now().Format("20060102150405"),
		Type:      eventType,
		Reason:    reason,
		Message:   message,
		Component: component,
		Timestamp: time.Now().UnixMilli(),
	}

	if eventType == EventTypeWarning {
		e.logger.Warn("Cluster Event", "reason", ev.Reason, "component", ev.Component, "msg", ev.Message)
	} else {
		e.logger.Info("Cluster Event", "reason", ev.Reason, "component", ev.Component, "msg", ev.Message)
	}
}
