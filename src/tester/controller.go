package main

import (
	"fmt"
	"golang.org/x/time/rate"
	"time"
	"context"
	
	Log "github.com/arush15june/load-tester/src/pkg/logger"
)

// messageRoutine follows the timing algorithm for sending a constant rate of messages.
func messageRoutine(hostname string, port string, sinktype string, routineConfig *MessageRoutineConfig) {

	localMsgRateVar := &routineConfig.MessageRate
	totalMsgSentVar := &routineConfig.MessageSent
	exitChan := routineConfig.ExitChan
	
	inititationWaiter.Add(1)
	finishWaiter.Add(1)

	var sendErr error
	var sentMsgAmt uint64

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
	
	limit    := rate.Every(MessageTime)
	limiter  := rate.NewLimiter(limit, 1)
	ctx      := context.Background()

	inititationWaiter.Done()
	inititationWaiter.Wait()

	secondRateTime := time.NewTicker(1 * time.Second)
	for {
		// Wait for rate limiter to signal start again.

		select {
		case <-secondRateTime.C:
			*localMsgRateVar = sentMsgAmt
			sentMsgAmt = 0
		case <-exitChan:
			return
		default:
		}
		limiter.WaitN(ctx, 1)
		_, sendErr = sinkConnection.SendPayload(Payload)
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
func deploySink(hostname string, port string, sinktype string, routineConfig *MessageRoutineConfig) {
	go messageRoutine(hostname, port, sinktype, routineConfig)
}