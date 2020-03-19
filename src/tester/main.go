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
	"os"
	"log"
	"runtime/pprof"
	"sync"
	"time"

	Log "github.com/arush15june/load-tester/src/pkg/logger"
	Args "github.com/arush15june/load-tester/src/tester/args"
)

// Global Variables
type MessageRoutineConfig struct {
	MessageRate uint64
	MessageSent uint64
	ExitChan chan bool
}

var (
	finishWaiter      sync.WaitGroup
	inititationWaiter sync.WaitGroup
	MessageRoutineList []*MessageRoutineConfig
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

		var sum uint64
		for n := range MessageRoutineList {
			sum += MessageRoutineList[n].MessageRate
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

	MessageRoutineList = make([]*MessageRoutineConfig, totalDevices)
	for i := range MessageRoutineList {
		routineConf := new(MessageRoutineConfig)
		routineConf.ExitChan = make(chan bool)
		MessageRoutineList[i] = routineConf
	}
	// msgRateArr = make([]uint64, totalDevices)
	// totalMsgSentArr = make([]uint64, totalDevices)
	// exitChanList = make([]chan bool, totalDevices)
	// for i := range exitChanList {
	// 	exitChanList[i] = make(chan bool)
	// }

	Log.LogInfo(fmt.Sprintf("Messages per device per second: %f/s\n", *Args.Messages))
	Log.LogInfo(fmt.Sprintf("Total messages per second: %f/s\n", totalMessagePerSecond))
	Log.LogInfo(fmt.Sprintf("Message Delta: %v\n", MessageTime))
	Log.LogInfo(fmt.Sprintf("Payload Size: %v bytes\n", *Args.PayloadSize))
	Log.LogInfo(fmt.Sprintf("Run Timer: %v\n", *Args.Timer))
	Log.LogInfo(fmt.Sprintf("Sink: %v\n", *Args.SinkType))

	Log.LogInfo(fmt.Sprintf("Ramping up %d workers.\n", totalDevices))
	if !*Args.NoExecute {

		// Setup Sinks and Listeners
		for i := range MessageRoutineList {
			deploySink(*Args.Hostname, *Args.Port, *Args.SinkType, MessageRoutineList[i])
			// deployListener()
		}

		inititationWaiter.Wait()
		Log.LogInfo(fmt.Sprintf("Started %d workers.\n", totalDevices))
	
		updateExitSignalChan := make(chan bool)
		go logMessageRate(updateExitSignalChan, &msgLogUpdateRate)
	
		Log.LogDebug(fmt.Sprintf("Timer: %v", *Args.Timer))
		if *Args.Timer >= 1*time.Second {
			exitTimer := time.NewTimer(*Args.Timer)
			select {
			case <-exitTimer.C:
				updateExitSignalChan <- true
				for routineIndex := range MessageRoutineList {
					MessageRoutineList[routineIndex].ExitChan <- true
				}
				Log.LogInfo("Timer finish.")
			}
		}
	
		finishWaiter.Wait()
	
		var msgSum uint64
		for i := range MessageRoutineList {
			msgSum += MessageRoutineList[i].MessageSent
		}
		Log.LogInfo(fmt.Sprintf("Total messages sent: %v\n", msgSum))
	}

}
