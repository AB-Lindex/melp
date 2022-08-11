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
	Validate() []error
	Connect() (Producer, error)
	Send(Message) (interface{}, error)
	Close() error
	Authorize(*http.Request) (bool, error)
}

// Consumer is the definition of something reveiving messages
type Consumer interface {
	Name() string
	Validate() []error
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

func (msg *Message) AddMetadata(key, value string) {
	if msg.Metadata == nil {
		msg.Metadata = make(map[string]string)
	}
	msg.Metadata[key] = value
}

func (msg *Message) AddHeader(key, value string) {
	if key == "" || value == "" {
		return
	}
	if msg.Headers == nil {
		msg.Headers = make(map[string]string)
	}
	msg.Headers[key] = value
}

func (msg *Message) ContentType() string {
	return ""
	// TODO
}
