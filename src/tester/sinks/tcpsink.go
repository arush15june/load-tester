package sinks

import (
	"net"
	"time"
)

// TCPSink is a message sink for a Connection.
type TCPSink struct {
	Payload []byte
	Conn    net.Conn
}

func (t *TCPSink) String() string {
	return "TCP"
}

// InitiateConnection dials a new TCP Connection.
func (t *TCPSink) InitiateConnection(hostname string, port string) error {
	conn, err := net.Dial("tcp", hostname+":"+port)
	if err != nil {
		return err
	}

	t.Conn = conn
	return nil
}

// SendPayload sends the payload via the TCP link.
func (t *TCPSink) SendPayload(payload []byte) (time.Time, error) {
	_, err := t.Conn.Write(payload)
	return time.Now(), err
}

// CloseConnection closes the TCP link.
func (t *TCPSink) CloseConnection() error {
	err := t.Conn.Close()
	return err
}
