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
	"sync"
	"sync/atomic"
	"time"
)

// Logger
var (
	infoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	LogInfo    = func(data string) {
		infoLogger.Output(2, data)
	}

	warnLogger = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime)
	LogWarn    = func(data string) {
		warnLogger.Output(2, data)
	}

	debugLogger = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime)
	LogDebug    = func(data string) {
		if *Verbose {
			debugLogger.Output(2, data)
		}
	}
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
	var start time.Time
	var sentMsgAmt uint64
	var exitVal = false

	sinkConnection, err := newSinkConnection(hostname, port, sinktype)
	if err != nil {
		LogDebug(fmt.Sprintf("Failed initiating %v connection to %v:%v", sinkConnection, hostname, port))
		inititationWaiter.Done()
		finishWaiter.Done()

		return
	}

	LogDebug(fmt.Sprintf("Initiated new %v connection to %v:%v", sinkConnection, hostname, port))

	secondRateTime := time.NewTimer(1 * time.Second)

	inititationWaiter.Done()
	inititationWaiter.Wait()
	for {
		select {
		case <-time.After(MessageTime - time.Since(start)):
		case <-secondRateTime.C:
			atomic.StoreUint64(localMsgRateVar, sentMsgAmt)
			sentMsgAmt = 0
			secondRateTime.Reset(1 * time.Second)
		case exitVal = <-exitChan:
			if exitVal {
				break
			}
		}

		start = time.Now()
		sendErr = sinkConnection.SendPayload(Payload)
		if sendErr != nil {
			fmt.Println(sendErr.Error())
			atomic.StoreUint64(localMsgRateVar, 0)
			break
		}

		sentMsgAmt++
		atomic.AddUint64(totalMsgSentVar, 1)
	}
	LogDebug(fmt.Sprintf("Closing %v connection to %v:%v\n", sinkConnection, hostname, port))
	sinkConnection.CloseConnection()

	finishWaiter.Done()
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

		if exitVal {
			break
		}

		var sum uint64 = 0
		for n, _ := range msgRateArr {
			sum += atomic.LoadUint64(&msgRateArr[n])
		}
		infoLogger.Printf("Message Rate: %v msg/s", sum)

		time.Sleep(*msgLogUpdateRate)
	}
}

func main() {
	ParseFlags()

	totalDevices := *Devices
	totalMessagePerSecond := *Messages * float64(totalDevices)
	msgLogUpdateRate := *UpdateRate

	MessageTime = getMessageTimeDuration(*Messages)
	Payload = generatePayload(*PayloadSize)
	msgRateArr = make([]uint64, totalDevices)
	totalMsgSentArr = make([]uint64, totalDevices)
	exitChanList = make([]chan bool, totalDevices)
	for i := range exitChanList {
		exitChanList[i] = make(chan bool)
	}

	LogInfo(fmt.Sprintf("Messages per device per second: %f/s\n", *Messages))
	LogInfo(fmt.Sprintf("Total messages per second: %f/s\n", totalMessagePerSecond))
	LogInfo(fmt.Sprintf("Message Delta: %v\n", MessageTime))
	LogInfo(fmt.Sprintf("Payload Size: %v bytes\n", *PayloadSize))
	LogInfo(fmt.Sprintf("Run Timer: %v\n", *Timer))

	LogInfo(fmt.Sprintf("Ramping up %d workers.\n", totalDevices))
	if !*NoExecute {
		for i := range exitChanList {
			routineMsgRateVar := &msgRateArr[i]
			routineExitChanVar := exitChanList[i]
			routineMsgSentVar := &totalMsgSentArr[i]
			deploySink(*Hostname, *Port, *SinkType, routineMsgRateVar, routineMsgSentVar, routineExitChanVar)
		}
	}

	inititationWaiter.Wait()
	LogInfo(fmt.Sprintf("Started %d workers.\n", totalDevices))

	updateExitSignalChan := make(chan bool)
	go logMessageRate(updateExitSignalChan, &msgLogUpdateRate)

	LogDebug(string(*Timer))
	if *Timer > 0*time.Second {
		exitTimer := time.NewTimer(*Timer)
		select {
		case <-exitTimer.C:
			updateExitSignalChan <- true
			for chanIndex := range exitChanList {
				exitChanList[chanIndex] <- true
			}
			LogInfo("Timer finish.")
		}
	}

	finishWaiter.Wait()

	var msgSum uint64 = 0
	for i := range totalMsgSentArr {
		msgSum += totalMsgSentArr[i]
	}
	LogInfo(fmt.Sprintf("Total messages sent: %v\n", msgSum))
}
