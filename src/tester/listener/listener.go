package listener

type Message struct {
	Payload []byte
	RecvdAt time.Time
}

// MessageListener is the interface used by message routines to receive messages from Pub/Sub topics.
type MessageListener interface {
	InitiateListener(hostname string, port string, topic string) error
	ReceiveMessage() Message, error
	CloseListener()
}