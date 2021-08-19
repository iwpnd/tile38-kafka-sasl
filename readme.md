# tile38-kafka-sasl

Test SASL authentification (SCRAM-SHA-512) between Tile38 1.25.2 and Apache Kafka.

## Prerequisites

`brew install tile38`

## Usage

`docker-compose up`

Once tile38 and kafka have successfully started use `tile38-cli` to pass two commands:

Setup a hook

```
sethook test kafka://broker:9091/test?auth=sasl&sha512=true within fleet fence object '{"type":"Feature","properties":{},"geometry":{"type":"Polygon","coordinates":[[[13.388900756835938,52.465422868400594],[13.423233032226562,52.465422868400594],[13.423233032226562,52.47964385776367],[13.388900756835938,52.47964385776367],[13.388900756835938,52.465422868400594]]]}}'
```

Set a test vehicle into the hook

```
set fleet test point 52.4723248048 13.4072685241
```

This will prompt you with:

```
tile38  | [sarama] Successfully initialized new client
tile38  | [sarama] ClientID is the default of 'sarama', you should consider setting it to something application-specific.
tile38  | [sarama] producer/broker/1001 starting up
tile38  | [sarama] producer/broker/1001 state change to [open] on test/0
tile38  | [sarama] Successful SASL handshake. Available mechanisms: [SCRAM-SHA-512]
tile38  | [sarama] SASL authentication succeeded
tile38  | [sarama] Connected to broker at broker:9091 (registered as #1001)
tile38  | 2021/07/25 16:06:10 [DEBU] Endpoint send ok: 5: kafka://broker:9091/auth=sasl&sha512=true: <nil>?
```
