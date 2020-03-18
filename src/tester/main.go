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
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"sync"
	"time"

	Log "github.com/arush15june/load-tester/src/pkg/logger"
	Args "github.com/arush15june/load-tester/src/tester/args"
)

// Global Variables
var (
	finishWaiter      sync.WaitGroup
	inititationWaiter sync.WaitGroup
	msgRateArr        []uint64
	totalMsgSentArr   []uint64
	exitChanList      []chan bool
)

// Tester is the container for load tester state.
type Tester struct {
	// State Variables

	// finishWaiter is the WaitGroup for completion of message goroutines.
	finishWaiter sync.WaitGroup

	// initiationWaiter is the WaitGroup for the ramp up of connections in message goroutines.
	inititationWaiter sync.WaitGroup

	// MsgRateArr is the array of message rates of individual goroutines.
	MsgRateArr []uint64

	// TotalMsgSentArr is the array containing total counts of messages sent by individual goroutines.
	TotalMsgSentArr []uint64

	// ExitChanList is the list of channels for signaling individual goroutines.
	ExitChanList []chan bool

	// Parameters

	// totalDevices is the total amount of concurrent connecting goroutines.
	totalDevices int

	// Hostname is the hostname of the sink
	Hostname string

	// Port is the port of the sink.
	Port string

	// Timer is the duration to run the tester for.
	Timer time.Duration

	// UpdateRate is the update rate for printing log messages.
	UpdateRate time.Duration

	// PayloadSize is the size of the payload.
	PayloadSize int

	// Derived Variables

	// MessageTime is the time for which each itearr
	MessageTime time.Duration

	// Payload is the payload of the load tester.
	Payload []byte

	// totalMessagePerSecond is the total messages per second guranteed to be sent by all devices.
	totalMessagePerSecond int
}

// Derived variables.
var (
	// MessageTime is T in the algorithm.
	MessageTime time.Duration

	// Payload is the generated payload bytetstream.
	Payload []byte
)

// messageRoutine follows the timing algorithm for sending a constant rate of messages.
func messageRoutine(hostname string, port string, sinktype string, localMsgRateVar *uint64, totalMsgSentVar *uint64, exitChan chan bool) {
	inititationWaiter.Add(1)
	finishWaiter.Add(1)

	var sendErr error
	var sentMsgAmt uint64
	var exitVal = false

	sinkConnection, err := newSinkConnection(hostname, port, sinktype)
	if err != nil {
		Log.LogDebug(fmt.Sprintf("Failed initiating %v connection to %v:%v", sinkConnection, hostname, port))
		Log.LogDebug(err.Error())
		inititationWaiter.Done()
		finishWaiter.Done()

		return
	}
	// Works on every return of function.
	defer func() {
		Log.LogDebug(fmt.Sprintf("Closing %v connection to %v:%v\n", sinkConnection, hostname, port))
		sinkConnection.CloseConnection()
		finishWaiter.Done()
	}()

	Log.LogDebug(fmt.Sprintf("Initiated new %v connection to %v:%v", sinkConnection, hostname, port))

	inititationWaiter.Done()
	inititationWaiter.Wait()

	var start time.Time
	secondRateTime := time.NewTicker(1 * time.Second)
	for {
		// Wait for rate limiter to signal start again.

		if *Args.Messages < 1 {
			select {
			case <-secondRateTime.C:
				*localMsgRateVar = sentMsgAmt
				sentMsgAmt = 0
			case exitVal = <-exitChan:
				if exitVal {
					return
				}
			default:
			}

		} else {
			select {
			case <-secondRateTime.C:
				*localMsgRateVar = sentMsgAmt
				sentMsgAmt = 0
			case exitVal = <-exitChan:
				if exitVal {
					return
				}
			case <-time.After(MessageTime - time.Since(start)):
			}
		}

		start = time.Now()
		sendErr = sinkConnection.SendPayload(Payload)
		if sendErr != nil {
			Log.LogFatal(sendErr.Error())
			*localMsgRateVar = 0
			break
		}

		sentMsgAmt++
		*totalMsgSentVar++
	}
}

// deploySink adds an entry to the WaitGroup and deploys the selected sink.
func deploySink(hostname string, port string, sinktype string, localMsgRateVar *uint64, totalMsgSentVar *uint64, routineExitChanVar chan bool) {
	go messageRoutine(hostname, port, sinktype, localMsgRateVar, totalMsgSentVar, routineExitChanVar)
}

func logMessageRate(signalChan chan bool, msgLogUpdateRate *time.Duration) {
	var exitVal = false
	for {
		select {
		case exitVal = <-signalChan:
		default:
		}

		if exitVal == true {
			return
		}

		Log.LogDebug(fmt.Sprintf("Messages Rate: %d", msgRateArr))
		var sum uint64
		for n := range msgRateArr {
			sum += msgRateArr[n]
		}
		Log.InfoLogger.Printf("Message Rate: %v msg/s", sum)

		time.Sleep(*msgLogUpdateRate)
	}
}

func main() {
	Args.ParseFlags()

	if *Args.Cpuprofile != "" {
		f, err := os.Create(*Args.Cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	totalDevices := *Args.Devices
	totalMessagePerSecond := *Args.Messages * float64(totalDevices)
	msgLogUpdateRate := *Args.UpdateRate

	MessageTime = getMessageTimeDuration(*Args.Messages)
	Payload = generatePayload(*Args.PayloadSize)
	msgRateArr = make([]uint64, totalDevices)
	totalMsgSentArr = make([]uint64, totalDevices)
	exitChanList = make([]chan bool, totalDevices)
	for i := range exitChanList {
		exitChanList[i] = make(chan bool)
	}

	Log.LogInfo(fmt.Sprintf("Messages per device per second: %f/s\n", *Args.Messages))
	Log.LogInfo(fmt.Sprintf("Total messages per second: %f/s\n", totalMessagePerSecond))
	Log.LogInfo(fmt.Sprintf("Message Delta: %v\n", MessageTime))
	Log.LogInfo(fmt.Sprintf("Payload Size: %v bytes\n", *Args.PayloadSize))
	Log.LogInfo(fmt.Sprintf("Run Timer: %v\n", *Args.Timer))
	Log.LogInfo(fmt.Sprintf("Sink: %v\n", *Args.SinkType))

	Log.LogInfo(fmt.Sprintf("Ramping up %d workers.\n", totalDevices))
	if !*Args.NoExecute {
		for i := range exitChanList {
			routineMsgRateVar := &msgRateArr[i]
			routineExitChanVar := exitChanList[i]
			routineMsgSentVar := &totalMsgSentArr[i]
			deploySink(*Args.Hostname, *Args.Port, *Args.SinkType, routineMsgRateVar, routineMsgSentVar, routineExitChanVar)
		}
	}

	inititationWaiter.Wait()
	Log.LogInfo(fmt.Sprintf("Started %d workers.\n", totalDevices))

	updateExitSignalChan := make(chan bool)
	go logMessageRate(updateExitSignalChan, &msgLogUpdateRate)

	Log.LogDebug(fmt.Sprintf("Timer: %v", *Args.Timer))
	if *Args.Timer > 0*time.Second {
		exitTimer := time.NewTimer(*Args.Timer)
		select {
		case <-exitTimer.C:
			updateExitSignalChan <- true
			for chanIndex := range exitChanList {
				exitChanList[chanIndex] <- true
			}
			Log.LogInfo("Timer finish.")
		}
	}

	finishWaiter.Wait()

	var msgSum uint64
	for i := range totalMsgSentArr {
		msgSum += totalMsgSentArr[i]
	}
	Log.LogInfo(fmt.Sprintf("Total messages sent: %v\n", msgSum))
}
