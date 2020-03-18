package sinks

import (
	"fmt"
	"math/rand"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	MqttDefaultPassword = "admin"
	MqttDefaultUsername = "admin"
	MqttDefaultClientId = "admin"
	MqttClientIdLength  = 10

	MqttDefaultCharset = "abcdefghijklmnopqrstuvwxyz"

	MqttDefaultTopic = "test"

	MqttConnectionTimeout = 3
)

// MQTTSink is a message sink for a MQTT Connection.
type MQTTSink struct {
	Payload []byte
	Client  mqtt.Client
}

func (t *MQTTSink) String() string {
	return "MQTT"
}

func getRandomId(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = MqttDefaultCharset[rand.Intn(len(MqttDefaultCharset))]
	}
	return string(b)
}

func mqttClientOptions(hostname string, port string) *mqtt.ClientOptions {
	clientOptions := mqtt.NewClientOptions()

	connURL := fmt.Sprintf("tcp://%s:%s", hostname, port)
	clientOptions.AddBroker(connURL)

	clientOptions.SetUsername(MqttDefaultUsername)
	clientOptions.SetPassword(MqttDefaultPassword)
	clientId := getRandomId(MqttClientIdLength)
	clientOptions.SetClientID(clientId)

	return clientOptions
}

// InitiateConnection connects a new MQTT Connection.
func (t *MQTTSink) InitiateConnection(hostname string, port string) error {
	clientOptions := mqttClientOptions(hostname, port)

	client := mqtt.NewClient(clientOptions)
	token := client.Connect()

	for !token.WaitTimeout(MqttConnectionTimeout * time.Second) {
	}

	if err := token.Error(); err != nil {
		return err
	}

	t.Client = client

	return nil
}

// SendPayload sends the payload via the MQTT link.
func (t *MQTTSink) SendPayload(payload []byte) error {
	token := t.Client.Publish(
		MqttDefaultTopic,
		0,
		false,
		payload,
	)
	token.Wait()
	if err := token.Error(); err != nil {
		return err
	}
	return nil
}

// CloseConnection closes the MQTT link.
func (t *MQTTSink) CloseConnection() error {
	t.Client.Disconnect(1)
	return nil
}
