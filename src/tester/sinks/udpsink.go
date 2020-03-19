package sinks

import (
	"net"
	"time"
)

// UDPSink is a message sink for a Connection.
type UDPSink struct {
	Payload []byte
	Conn    net.Conn
}

func (t *UDPSink) String() string {
	return "UDP"
}

// InitiateConnection dials a new UDP Connection.
func (t *UDPSink) InitiateConnection(hostname string, port string) error {
	conn, err := net.Dial("udp", hostname+":"+port)
	if err != nil {
		return err
	}

	t.Conn = conn
	return nil
}

// SendPayload sends the payload via the UDP link.
func (t *UDPSink) SendPayload(payload []byte) (time.Time, error) {
	_, err := t.Conn.Write(payload)
	return time.Now(), err
}

// CloseConnection closes the UDP link.
func (t *UDPSink) CloseConnection() error {
	err := t.Conn.Close()
	return err
}
