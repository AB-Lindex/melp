package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

var passthruHeaders = []string{
	"Content-Type",
	"Content-Encoding",
	"Content-Disposition",
	"Content-Language",
	"Content-Range",
	"ETag",
	"Expires",
	"X-Request-Id",
	"X-Correlation-Id",
	"Traceparent",
	"Tracestate",
}

type stringError string

func (err stringError) Error() string {
	return string(err)
}

type requiredError string

func (name requiredError) Error() string {
	return fmt.Sprintf("'%s' is required", string(name))
}

type invalidError string

func (name invalidError) Error() string {
	return fmt.Sprintf("'%s' is invalid", string(name))
}

// Producer is the definition of something that send messages
type Producer interface {
	Name() string
	Validate() ([]error, bool)
	Connect() (Producer, error)
	Send(Message, *http.Request) (interface{}, error)
	Close() error
	Authorize(*http.Request) (bool, error)
}

// Consumer is the definition of something reveiving messages
type Consumer interface {
	Name() string
	Validate() ([]error, bool)
	Connect() (Consumer, error)
	Listen(*sync.WaitGroup)
	Close() error
}

// Message is the data that is sent or received
type Message struct {
	Body      []byte
	Timestamp time.Time
	Metadata  map[string]string
	Headers   map[string]string
}

// AddMetadata adds metadata (key+value) to the message
func (msg *Message) AddMetadata(key, value string) {
	if msg.Metadata == nil {
		msg.Metadata = make(map[string]string)
	}
	msg.Metadata[key] = value
}

// AddHeader adds header(s) (key+value) to the message,
// all keys are converted using http.CanonicalHeaderKey
func (msg *Message) AddHeader(key, value string) {
	if key == "" || value == "" {
		return
	}
	if msg.Headers == nil {
		msg.Headers = make(map[string]string)
	}
	msg.Headers[http.CanonicalHeaderKey(key)] = value
}

// GetHeader returns the header value of a specific key
// (the key is converted using http.CanonicalHeaderKey before match)
func (msg *Message) GetHeader(key string) string {
	return msg.Headers[http.CanonicalHeaderKey(key)]
}

// ContentType returns the content-type of the message
func (msg *Message) ContentType() string {
	return msg.Headers["Content-Type"]
}
