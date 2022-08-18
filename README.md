# melp
A Message Helper to proxy between messages and REST-APIs

Please check out the usage-guide with [REST-API](./REST-API.md) examples

## Send message using REST-API (output)

This setup is when the `Actor` is unable to directly talk to the desired messaging system (firewall or 3rd-party product are common scenarions).

With this setup the `actor` can instead do a HTTP POST to `melp` which will translate that into an actual message to the message-broker.

```mermaid
flowchart LR;
    A([Actor])
    O[[Melp]]
    C([Messaging System])
    A== "1) POST /send/ID" ==>O
    O== "2) (msg protocol)" ==>C
    C-- "3) (ok)" -->O
    O-- "4) 200 OK" -->A
```

[More information..](./docs/REST-API.md)

## Receive message (like a webhook)

```mermaid
flowchart LR;
    R([Receiver])
    O[[Melp]]
    M([Messaging System])
    M== "1) Message" ==>O
    O== "2) POST /custom/url" ==>R
    R-- "3) 200 Ok" -->O
    O-- "4) (ok)" -->M
```
    