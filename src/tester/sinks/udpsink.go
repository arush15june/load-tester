package sinks

import "net"

// UDPSink is a message sink for a Connection.
type UDPSink struct {
	Conn net.Conn
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
func (t *UDPSink) SendPayload(payload []byte) error {
	_, err := t.Conn.Write(payload)
	return err
}

// CloseConnection closes the UDP link.
func (t *UDPSink) CloseConnection() error {
	err := t.Conn.Close()
	return err
}
