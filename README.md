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

# Usage
```bash
go build src/main.go
./main -h
```

## Flags
- **-devices** int
    - Number of devices/connections. (default 1)
- **-hostname** string
    - Hostname of sink. (default "127.0.0.1")
- **-msg** float
    - Number of messages per second. (default 2)
- **-payload** int
    - Payload size in bytes. (default 64)
- **-port** string
    - Port of sink. (default "18000")
- **-sink** string
    - Sink required. [tcp, udp, mqtt] (default "tcp")
- **-noexec** bool
        Don't execute.

# TODO
- Better logging and metrics.
- More Sinks!
- Kafka Sink
- RabbitMQ Sink
- NATS Sink
- Research clock resolution, maximum feasible messages/second.
- Fixed duration testing.
- Real time message rate monitoring.
- Message/sec/routine limit acc to clock resolution.
- Better error handling (Graceful Connection Close).