# Send message using REST-API (output)

This setup is when the `Actor` is unable to directly talk to the desired messaging system (firewall or 3rd-party product are common scenarions).

With this setup the `actor` can instead do a HTTP POST to `melp` which will translate that into an actual message to the message-broker.

```mermaid
flowchart LR;
    A([Actor])
    O[[Melp]]
    C([Messaging System])
    A== POST /send/ID ==>O
    O-- 200 OK -->A
    O== "(msg protocol)" ==>C
    C-- "(ok)" -->O
``````

## Configuration

```yaml
output:
  <provider>:
  - id: ID_FOR_POST_URL
    ...provider-config...
    auth:
      anon: false
      bearer: BEARER_TOKEN_FROM_ACTOR
      basic:
        ACTOR_USERNAME1: ACTOR_PASSWORD1
        ACTOR_USERNAME2: ACTOR_PASSWORD2
```

The `id`' is used in the URL on how to access this specific output 

The following providers are supported:
* [Kafka](KAFKA.md)

## Authorization

### Outputs
All `output`s must have an `auth` section to either say that anonymous access is allowed, or what kind of HTTP Authorization is required.

Just add this to you `output` entry
```yaml
```

If you want to enable anonymous access to your `output` then add this instead
```yaml
    auth:
      anon: true
```
