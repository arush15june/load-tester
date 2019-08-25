package main

import (
	"flag"
	"time"
)

// Argument Flags
var (
	// Devices is the number of concurrent devices/connections
	// N in algorithm.
	Devices = flag.Int("devices", 1, "Number of devices/connections.")

	// Messages is the number of messages per second.
	// msg in algorithm.
	Messages = flag.Float64("msg", 2, "Number of messages per second.")

	// Hostname is the hostname of the sink
	Hostname = flag.String("hostname", "127.0.0.1", "Hostname of sink.")

	// Port is the port of the sink.
	Port = flag.String("port", "18000", "Port of sink.")

	// PayloadSize is the size of the payload to sink.
	PayloadSize = flag.Int("payload", 64, "Payload size in bytes.")

	// SinkType is the type of sink required on hostname:port.
	SinkType = flag.String("sink", "tcp", "Sink required. [tcp, udp, mqtt]")

	// Timer is the duration to run the tester for.
	Timer = flag.Duration("duration", 2*time.Second, "Duration to run for. 0 for inifite.")

	// Timer is the duration to run the tester for.
	NoExecute = flag.Bool("noexec", false, "Don't execute.")

	// UpdateRate is the update rate for printing log messages.
	UpdateRate = flag.Duration("update", 1*time.Second, "Message rate log frequency. Faster update might affect performance.")

	// Verbose enables verbose logging.
	Verbose = flag.Bool("verbose", false, "Verbose mode logging.")
)

// ParseFlags parses arg flags.
func ParseFlags() {
	flag.Parse()
}
