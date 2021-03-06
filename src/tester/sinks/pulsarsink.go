package sinks

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/apache/pulsar/pulsar-client-go/pulsar"
)

const (
	PulsarDefaultTopic = "loadtester"
)

// PulsarSink is a message sink for a NATS Subscription.
type PulsarSink struct {
	Payload  []byte
	Client   pulsar.Client
	Producer pulsar.Producer
}

func (t *PulsarSink) String() string {
	return "NATS"
}

// InitiateConnection starts a NATS Connection
func (t *PulsarSink) InitiateConnection(hostname string, port string) error {
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:                     fmt.Sprintf("pulsar://%s:%s", hostname, port),
		OperationTimeoutSeconds: 5,
		MessageListenerThreads:  runtime.NumCPU(),
	})

	if err != nil {
		return err
	}

	producer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic: PulsarDefaultTopic,
	})

	if err != nil {
		return err
	}

	t.Client = client
	t.Producer = producer

	return nil
}

// SendPayload publishes a NATS payload.
func (t *PulsarSink) SendPayload(payload []byte) (time.Time, error) {
	msg := pulsar.ProducerMessage{
		Payload: payload,
	}

	err := t.Producer.Send(context.Background(), msg)
	return time.Now(), err
}

// CloseConnection closes the NATS connection.
func (t *PulsarSink) CloseConnection() error {
	t.Client.Close()
	return nil
}
