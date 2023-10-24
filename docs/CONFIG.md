# Melp Config

## Overview
```yaml
apiVersion: melp/v1alpha1

endpoints:
  ...

producers:
  ...

consumers:
  ...
```

### Environment variables
Before parsing the config-file, all occurences of `${..}` are expanded using environment-variables.
Just write your (case-sensitive) environment-variable name between the curly-braces.

## Endpoints
```yaml
endpoints:
  kafka:
    - name: kafka-1
      endpoint: sasl_ssl://kafka.local:9092
      key: xyzzy
      secret: ACTUAL_SECRET_VALUE
```
This is a shared config for all producers and consumers, allowing multiple producers and consumers to share the same endpoint-settings.<br/>
The `name` is used to reference the endpoint from the producers and consumers.

### `endpoint`
The structure for the endpoint is `schema://fqdn:port`

This part is parsed as a series or keywords separated with an underscore ('_' character).
(keywords are case-insensitive)

| Keyword | Effect |
| ------ | ------- |
| sasl | Require `key` and `secret` for the SASL network connection (SASLTypePlaintext) |
| ssl | This will enable TLS on the connection |

## Producers
```yaml
producers:
  kafka:
    - endpoint: kafka-1  # reference to endpoint (see above)
      topic: my-topic-name-to-send
      id: URL_ID
      auth:
        anon: true
```
The `id` is used to reference which producer-topic to access, and is part of the URL to call:
`/send/URL_ID`

The `auth` section is required, and is used to specify authentication for the caller.

### `auth` examples:
You have to specify one of the following authentication methods:
```yaml
      auth:
        anon: true
```

```yaml
      auth: 
        bearer: LITERAL_VALUE_OF_LONG_BEARER_TOKEN
```

```yaml
      auth: 
        basic:
          username1: password1
          username2: password2
          ...
```

You can have both `basic` and `bearer` in the same `auth` section, any match will be accepted.

## Consumers
```yaml
consumers:
  kafka:
    - endpoint: kafka-1  # reference to endpoint (see above)
      group: kafka-consumer-group-name
      topics:
        - my-topic-name-to-read
      id: CONSUMER_ID
      callback:
        url: http://localhost:8080/callback/%{topic}
        headers:
          Customer-Http-Header: my-special-value
```

These are the consumers that will read messages from Kafka and send them to the callback-url.

All topics within one `consumer.kafka` will use the same consumer-group.

You can place the string `%{topic}` in the callback-url, and it will be replaced with the topic-name for each message.

The `headers` is an optional key/value-map of headers to add to the callback-request.