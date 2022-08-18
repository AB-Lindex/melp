# Kafka-specifics

The Kafka provider in `melp` is named `kafka`

Kafka-specific shared config:
```yaml
endpoint: sasl_ssl://kafka.local:9092
key: MY_API_KEY
secret: MY_API_SECRET
```

## Output
A Kafka output (producer) is sending message to a `topic`

Kafka-specific output values:
```yaml
topic: KAFKA_TOPIC
```

## Input
A Kafka input (consumer) is using a 'consumer-group' to read from one or more topics

Kafka-specific input values:
```yaml
group: CONSUMER_GROUP
topics:
- KAFKA_TOPIC
```

## Endpoint

The structure for the endpoint is `schema://fqdn:port`

## Schema
This part is parsed as a series or keywords separated with an underscore ('_' character).
(keywords are case-insensitive)

| Keyword | Effect |
| ------ | ------- |
| sasl | Require `key` and `secret` for the SASL network connection (SASLTypePlaintext) |
| ssl | This will enable TLS on the connection |


## Full example (of output)
```yaml
output:
  kafka:
  - id: ID_FOR_POST_URL
    endpoint: sasl_ssl://kafka.local:9092
    key: MY_API_KEY
    secret: MY_API_SECRET
    topic: KAFKA_TOPIC
    auth:
      anon: false
      bearer: BEARER_TOKEN_FROM_ACTOR
      basic:
        ACTOR_USERNAME1: ACTOR_PASSWORD1
        ACTOR_USERNAME2: ACTOR_PASSWORD2
```