# load-tester

Parallel load tester for networked services. Send a constant amount of messages per second across concurrent connections.

# Algorithm
```
- N = number of devices
- msg = number of messages per second.
- T = 1/msg -> Time taken by 1 message to guarantee msg msgs/second.
- Payload = Payload in memory. (bytes)
- For every device, Initialize a goroutine,
- Inside goroutine
  - Initiate Connection.
  - On Every Iteratioin
    - Start timing.
      - SendPayload(Payload)
    - Stop Timing => Tsend.
    - Sleep(T - Tsend)
```

## Design
- Metrics interface
  - Generic metrics interface.
  - Concrete implementations for service metrics like Round Trip Times, HTTP Request Times, etc.
  - metrics stored as key-values, exposed as mappings.
  - Thread safe implmenetations to update via LoadTesters goroutines.
  - serializable state, can also be used for comparison instead of holding state in LoadTester's.
Ex. HTTP Metrics, TCP Metrics, UDP Metrics, MQTT Metrics, AMQP Metrics, etc.

- LoadTester interface
  - Implement load testing, generate metrics, handle timing.
  - Concrete implementations for specific service testing and concrete metrics.
  - Encompass concrete metrics and calculate metrics.
  - Mediator in essence, for Sinks and Metrics.
Ex. HTTPTester, TCPTester, UDPTester, MQTTTester, AMQPTester, etc.

# Usage
```bash
go build ./src/tester
./tester -h
```

## Flags
  - **devices** int
      - Number of devices/connections. (default 1)
  - **duration** duration
      - Duration to run for. 0 for inifite. (default 2s)      
  - **hostname** string
      - Hostname of sink. (default "127.0.0.1")
  - **msg** float
      - Number of messages per second. (default 2)
  - **noexec**
      - Don't execute.
  - **payload** int
      - Payload size in bytes. (default 64)
  - **port** string
      - Port of sink. (default "18000")
  - **sink** string
      - Sink required. [tcp, udp, mqtt] (default "tcp")       
  - **update** duration
      - Message rate log frequency. Faster update might affect performance. (default 1s)
  - **verbose**
      - Verbose mode logging.

# TODO
- More Sinks!
- Kafka Sink
- RabbitMQ Sink
- NATS Sink
- LoadTester and Metrics architecture.
- Metrics for various services.
- Research clock resolution, maximum feasible messages/second.
- Message/sec/routine limit acc to clock resolution.
- Better error handling (dont panic on no connection)
- Distributed load testing.