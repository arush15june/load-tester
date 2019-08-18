package main

// Load Tester for Network Sinks.
//
// 	Model load via N devices (connections) with each sending N msg/s.
//
// Algorithm
// - N = number of devices
// - msg = number of messages per second.
// - T = 1/msg -> Time taken by 1 message to guarantee msg msgs/second.
// - Payload = Payload in memory. (bytes)
// - For every device, Initialize a goroutine,
// - Inside goroutine
// 		- Initiate Connection.
// 		- On Every Iteratioin
// 			- Start timing.
// 				- SendPayload(Payload)
// 			- Stop Timing => Tsend.
// 			- Sleep(T - Tsend)

import (
	"flag"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Global Variables
var (
	waiter     sync.WaitGroup
	msgRateArr []uint64
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
	Timer = flag.Duration("duration", 0, "Duration to run for. 0 for inifite.")

	// Timer is the duration to run the tester for.
	NoExecute = flag.Bool("noexec", false, "Don't execute.")

	// UpdateRate is the update rate for printing log messages.
	UpdateRate = flag.Duration("update", 100*time.Millisecond, "Message rate log rate. Faster update might affect performance.")
)

// Derived constants.
var (
	// MessageTime is T in the algorithm.
	MessageTime time.Duration

	// Payload is the generated payload bytetstream.
	Payload []byte
)

// messageRoutine follows the timing algorithm for sending a constant rate of messages.
func messageRoutine(hostname string, port string, sinktype string, localMsgRateVar *uint64) {
	defer waiter.Done()

	var sendErr error
	var start time.Time
	var sentMsgAmt uint64

	sinkConnection := newSinkConnection(hostname, port, sinktype)
	// fmt.Printf("Initiated new %v connection to %v:%v\n", sinkConnection, hostname, port)

	secondTimer := time.NewTimer(1 * time.Second)

	for {
		start = time.Now()
		sendErr = sinkConnection.SendPayload(Payload)
		if sendErr != nil {
			fmt.Println(sendErr.Error())
			atomic.StoreUint64(localMsgRateVar, 0)
			break
		}

		sentMsgAmt += 1
		select {
		case <-secondTimer.C:
			atomic.StoreUint64(localMsgRateVar, sentMsgAmt)
			sentMsgAmt = 0
			secondTimer.Reset(1 * time.Second)
		default:
		}

		time.Sleep(MessageTime - time.Since(start))
	}
	fmt.Printf("Closing %v connection to %v:%v\n", sinkConnection, hostname, port)
	sinkConnection.CloseConnection()
}

func deploySink(hostname string, port string, sinktype string, localMsgRateVar *uint64) {
	waiter.Add(1)
	go messageRoutine(hostname, port, sinktype, localMsgRateVar)
}

func main() {
	flag.Parse()

	totalDevices := *Devices
	totalMessagePerSecond := *Messages * float64(totalDevices)
	msgLogUpdateRate := *UpdateRate

	MessageTime = getMessageTimeDuration(*Messages)
	Payload = generatePayload(*PayloadSize)
	msgRateArr = make([]uint64, totalDevices)

	fmt.Printf("Messages per device per second: %f/s\n", *Messages)
	fmt.Printf("Total messages per second: %f/s\n", totalMessagePerSecond)
	fmt.Printf("Message Delta: %v\n", MessageTime)
	fmt.Printf("Payload Size: %v bytes\n", *PayloadSize)

	if !*NoExecute {
		for i := 0; i < totalDevices; i++ {
			routineMsgRateVar := &msgRateArr[i]
			deploySink(*Hostname, *Port, *SinkType, routineMsgRateVar)
		}
	}

	go func() {
		for {
			var sum uint64 = 0
			for _, num := range msgRateArr {
				sum += num
			}
			fmt.Print("Message Rate: ")
			fmt.Printf("%v msg/s", sum)
			fmt.Printf("\r")
			time.Sleep(msgLogUpdateRate)
		}
	}()

	waiter.Wait()
}
