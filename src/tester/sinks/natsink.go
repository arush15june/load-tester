package sinks

import (
	"fmt"
	"time"
	nats "github.com/nats-io/nats.go"
)

const (
	NatsDefaultSubject = "loadtester"
)

// NATSSink is a message sink for a NATS Subscription.
type NATSSink struct {
	Payload []byte
	Client  *nats.Conn
}

func (t *NATSSink) String() string {
	return "NATS"
}

// InitiateConnection starts a NATS Connection
func (t *NATSSink) InitiateConnection(hostname string, port string) error {
	nc, err := nats.Connect(fmt.Sprintf("%s:%s", hostname, port))
	if err != nil {
		return err
	}

	t.Client = nc

	return nil
}

// SendPayload publishes a NATS payload.
func (t *NATSSink) SendPayload(payload []byte) (time.Time, error) {
	err := t.Client.Publish(NatsDefaultSubject, payload)
	return time.Now(), err
}

// CloseConnection closes the NATS connection.
func (t *NATSSink) CloseConnection() error {
	t.Client.Close()
	return nil
}
