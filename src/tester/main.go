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
		fmt.Println(data)
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
	var exitVal bool = false

	sinkConnection := newSinkConnection(hostname, port, sinktype)
	LogDebug(fmt.Sprintf("Initiated new %v connection to %v:%v\n", sinkConnection, hostname, port))

	secondRateTime := time.NewTimer(1 * time.Second)

	inititationWaiter.Done()
	inititationWaiter.Wait()

	for {
		select {
		case <-secondRateTime.C:
			atomic.StoreUint64(localMsgRateVar, sentMsgAmt)
			sentMsgAmt = 0
			secondRateTime.Reset(1 * time.Second)
		case exitVal = <-exitChan:
		default:
		}

		if exitVal {
			break
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

		time.Sleep(MessageTime - time.Since(start))
	}
	LogDebug(fmt.Sprintf("Closing %v connection to %v:%v\n", sinkConnection, hostname, port))
	sinkConnection.CloseConnection()

	finishWaiter.Done()
}

// deploySink adds an entry to the WaitGroup and deploys the selected sink.
func deploySink(hostname string, port string, sinktype string, localMsgRateVar *uint64, totalMsgSentVar *uint64, routineExitChanVar chan bool) {
	go messageRoutine(hostname, port, sinktype, localMsgRateVar, totalMsgSentVar, routineExitChanVar)
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
	go func(signalChan chan bool) {
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
			for _, num := range msgRateArr {
				sum += num
			}
			LogInfo(fmt.Sprintf("Message Rate: %v msg/s\n", sum))

			time.Sleep(msgLogUpdateRate)
		}
	}(updateExitSignalChan)

	exitTimer := time.NewTimer(*Timer)
	select {
	case <-exitTimer.C:
		updateExitSignalChan <- true
		for chanIndex := range exitChanList {
			exitChanList[chanIndex] <- true
		}
		LogInfo("Timer finish.")
	}

	finishWaiter.Wait()

	var msgSum uint64 = 0
	for i := range totalMsgSentArr {
		msgSum += totalMsgSentArr[i]
	}
	LogInfo(fmt.Sprintf("Total messages sent: %v\n", msgSum))
}
