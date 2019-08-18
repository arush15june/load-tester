package sinks

// MessageSink is the interface used by message routines to send messages.
type MessageSink interface {
	InitiateConnection(hostname string, port string) error
	SendPayload(payload []byte) error
	CloseConnection() error
}
