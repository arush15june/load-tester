package sinks

import (
	"net"
)

// MessageSink is the interface used by message routines to send messages.
type MessageSink interface {
	InitiateConnection(hostname string, port string) error
	SendPayload(payload []byte) error
	CloseConnection() error
}

// TCPSink is a message sink for a Connection.
type TCPSink struct {
	Conn net.Conn
}

// UDPSink is a message sink for UDP Connection.
type UDPSink struct {
	Conn net.Conn
}

func (t *TCPSink) InitiateConnection(hostname string, port string) error {
	conn, err := net.Dial("tcp", hostname+":"+port)
	if err != nil {
		return err
	}

	t.Conn = conn
	return nil
}

func (t *TCPSink) SendPayload(payload []byte) error {
	_, err := t.Conn.Write(payload)
	return err
}

func (t *TCPSink) CloseConnection() error {
	err := t.Conn.Close()
	return err
}

// func (t UDPSink) InitiateConnection(hostname string, port string) error {

// }

// func (t UDPSink) SendPayload(payload []byte) error {

// }

// func (t UDPSink) CloseConnection() error {

// }
