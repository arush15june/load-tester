package main

// Load Tester for Network Sinks.
//
// 	Model load via N devices (connections) with each sending N msg/s.
//
// 	Algorithm
// 	- N = number of devices
// 	- msg = number of messages per second.
// 	- T = 1/msg -> Time taken by 1 message to guarantee msg msgs/second.
// 	- Payload = Payload in memory. (bytes)
// 	- For every device, Initialize a goroutine,
// 	- Inside goroutine
// 		- Initiate Connection.
//  	- On Every Iteratioin
// 			- Start timing.
// 				- SendPayload(Payload)
// 			- Stop Timing => Tsend.
// 			- Sleep(T - Tsend)

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	Sinks "github.com/arush15june/load-tester/src/sinks"
)

// Constants
const (
	PayloadSymbol = "A"
)

// Global Variables
var (
	waiter sync.WaitGroup
)

// Argument Flags
var (
	// Devices is the number of concurrent devices/connections
	// T in algorithm.
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
)

// Derived constants.
var (
	// MessageTime is T in the algorithm.
	MessageTime time.Duration

	// Payload is the generated payload bytetstream.
	Payload []byte
)

func getMessageTimeDuration(msgs float64) time.Duration {
	T := 1.0 / msgs
	timeString := strconv.FormatFloat(T, 'g', -1, 64) + "s"

	duration, _ := time.ParseDuration(timeString)
	return duration
}

func generatePayload(size int) []byte {
	return []byte(strings.Repeat(PayloadSymbol, size-1) + "\n")
}

func newSinkConnection(hostname string, port string) Sinks.MessageSink {
	sinkConn := new(Sinks.TCPSink)
	sinkConn.InitiateConnection(hostname, port)

	return sinkConn
}

func messageRoutine(hostname string, port string) {
	defer waiter.Done()

	fmt.Printf("Iniating new connection to %v:%v\n", hostname, port)
	sinkConnection := newSinkConnection(hostname, port)

	for {
		start := time.Now()
		sinkConnection.SendPayload(Payload)

		time.Sleep(MessageTime - time.Since(start))
	}
	sinkConnection.CloseConnection()
}

func main() {
	flag.Parse()

	MessageTime = getMessageTimeDuration(*Messages)
	Payload = generatePayload(*PayloadSize)

	fmt.Printf("Messages per second: %fs\n", *Messages)
	fmt.Printf("Message Delta: %v\n", MessageTime)
	fmt.Printf("Paylod Size: %d bytes\n", PayloadSize)

	for i := 0; i < *Devices; i++ {
		waiter.Add(1)
		go messageRoutine(*Hostname, *Port)
	}

	waiter.Wait()
}
