package sinks

import (
	"time"
)

// MessageSink is the interface used by message routines to send messages.
type MessageSink interface {
	InitiateConnection(hostname string, port string) error
	SendPayload(payload []byte) (time.Time, error)
	CloseConnection() error
}
