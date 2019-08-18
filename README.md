# load-tester

Parallel load tester for networked services. Send a cosntant amount of messages per second accros concurrent connections.

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

# TODO
- Better logging and metrics.
- More Sinks!
- UDP Sink
- MQTT Sink
- Kafka Sink
- RabbitMQ Sink
- NATS Sink