# melp REST-API

## Send message

You do a `POST` to `/send/{topic-address}` with the body as you data

You can send describing headers with the message (see below)

Usually you also need to authorize yourself using the `Authorization` header.
It can be either a `Bearer` or `Basic` requirement

```sh
curl --request POST \
  --url 'http://HOSTNAME/send/OUTPUT-ID' \
  --header 'Authorization: Bearer YOUR_BEARER_TOKEN' \
  --header 'Content-Type: application/json' \
  --data '{"data":"text"}'
```

## Receive message

When you're receive messages you will get a request that will look something like this:

```
POST /your/path?topic=MESSAGE_TOPIC
Content-Type: application/json
Content-Length: 15

{"data":"text"}
```

| Returned Status Code | Action taken |
| -------------------- | ------------ |
| 200 Ok               | Message accepted |
| 404 Not Found        | Not marked as processed |
| 5xx Errors           | Retry up to 5 times, then not marked as processed |

If you want (or need) to require authentication `melp` can send either `Bearer` or `Basic` authorization.

### Message not marked as processed

These will be resent at a later time.

When `Kafka` is used this might cause processed messages after a failed on to be resent, thus being duplicates. Please use the Headers `Melp-Partition` and `Melp-Offset` to uniquely track each message.

## Headers
The following headers can be passed with the message:
| Header           | Example                           |
| ---------------- | --------------------------------- |
| Content-Type     | application/json, application/xml |
| Content-Encoding | gzip                              |
| Content-Language | en-US                             |
