package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/ninlil/butler/log"
)

// Misc global text-constants
var (
	ID       = "id"
	ENDPOINT = "endpoint"
	KEY      = "key"
	SECRET   = "secret"
	TOPIC    = "topic"
	TOPICS   = "topics"
	GROUP    = "group"
	URL      = "callback.url"
	//CALLBACK = "callback"
	//AUTH     = "auth"

	errEndpointMissingHost   = stringError("'endpoint' is missing host and/or port")
	errEndpointInvalidScheme = stringError("'endpoint' is invalid")
)

type kafkaEndpoint struct {
	endpoint string
	key      string
	secret   string

	err error

	scheme string
	host   string
	port   int
	ssl    bool
	sasl   bool
}

type saramaLogger struct{}

func (slog *saramaLogger) Write(txt string) {
	txt = strings.TrimRight(txt, "\n\r")
	log.Trace().Msgf("sarama: %s", txt)
}

func (slog *saramaLogger) Print(v ...interface{}) {
	slog.Write(fmt.Sprint(v...))
}

func (slog *saramaLogger) Println(v ...interface{}) {
	slog.Write(fmt.Sprint(v...))
}

func (slog *saramaLogger) Printf(format string, args ...interface{}) {
	slog.Write(fmt.Sprintf(format, args...))
}

func initKafka() {
	sarama.Logger = new(saramaLogger)
}

func newKafkaEndpoint(endpoint, key, secret string) *kafkaEndpoint {
	var e = &kafkaEndpoint{
		endpoint: os.ExpandEnv(endpoint),
		key:      os.ExpandEnv(key),
		secret:   os.ExpandEnv(secret),
	}

	e.err = e.parseEndpoint()

	return e
}

func (ep *kafkaEndpoint) parseEndpoint() error {
	var hostport, port string

	parts := strings.SplitN(ep.endpoint, "://", 2)
	switch len(parts) {
	case 0:
		return requiredError(ENDPOINT)
	case 1:
		hostport = parts[0]
	case 2:
		ep.scheme = parts[0]
		hostport = parts[1]
	}

	parts = strings.SplitN(hostport, ":", 2)
	switch len(parts) {
	case 0:
		return requiredError(ENDPOINT)
	case 1:
		ep.host = parts[0]
	case 2:
		ep.host = parts[0]
		port = parts[1]
	}
	if port == "" {
		ep.port = 9092
	} else {
		i, err := strconv.Atoi(port)
		if err != nil {
			return err
		}
		ep.port = i
	}

	if ep.host == "" || ep.port == 0 {
		return errEndpointMissingHost
	}

	parts = strings.Split(strings.ToUpper(ep.scheme), "_")
	for _, scheme := range parts {
		switch scheme {
		case "SASL":
			ep.sasl = true
		case "SSL":
			ep.ssl = true
		default:
			return errEndpointInvalidScheme
		}
	}

	return nil
}

func (ep *kafkaEndpoint) SetConfig(cfg *sarama.Config) {
	if ep.sasl {
		cfg.Net.SASL.Enable = true
		cfg.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		cfg.Net.SASL.User = ep.key
		cfg.Net.SASL.Password = ep.secret
	}

	if ep.ssl {
		cfg.Net.TLS.Enable = true
	}
}

func (ep *kafkaEndpoint) Peers() []string {
	return []string{fmt.Sprintf("%s:%d", ep.host, ep.port)}
}

func (ep *kafkaEndpoint) Error() error {
	return ep.err
}

func (ep *kafkaEndpoint) Validate() error {

	if ep.err != nil {
		return ep.err
	}

	if ep.endpoint == "" {
		return requiredError(ENDPOINT)
	}

	if ep.sasl {
		if ep.key == "" {
			return requiredError(KEY)
		}
		if ep.secret == "" {
			return requiredError(SECRET)
		}
	}
	return nil
}
