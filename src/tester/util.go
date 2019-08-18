package main

import (
	"strconv"
	"strings"
	"time"

	Sinks "github.com/arush15june/load-tester/src/tester/sinks"
)

// Constants
const (
	PayloadSymbol = "A"
)

// getMessageTimeDuration returns the duration of message derived from required msgs/second.
func getMessageTimeDuration(msgs float64) time.Duration {
	T := 1.0 / msgs
	timeString := strconv.FormatFloat(T, 'f', -1, 64) + "s"

	duration, _ := time.ParseDuration(timeString)
	return duration
}

// generatePayload generate the required message sized payload.
func generatePayload(size int) []byte {
	return []byte(strings.Repeat(PayloadSymbol, size-1) + "\n")
}

// newSinkConnnection creates a new Sinks.MessageSink
func newSinkConnection(hostname string, port string, typ string) Sinks.MessageSink {
	var sinkConn Sinks.MessageSink
	if typ == "tcp" {
		sinkConn = new(Sinks.TCPSink)
	} else if typ == "udp" {
		sinkConn = new(Sinks.UDPSink)
	} else if typ == "mqtt" {
		sinkConn = new(Sinks.MQTTSink)
	}

	sinkConn.InitiateConnection(hostname, port)

	return sinkConn
}
