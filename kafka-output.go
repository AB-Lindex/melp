package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Shopify/sarama"
	"github.com/ninlil/butler/log"
)

type kafkaProducer struct {
	msg       sarama.SyncProducer
	connected bool

	ID       string
	Topic    string
	Endpoint *kafkaEndpoint

	Auth *Auth
}

func (p *kafkaProducer) Name() string {
	return p.ID
}

func (p *kafkaProducer) Validate() ([]error, bool) {
	var errs []error

	if p.Endpoint.err != nil {
		errs = append(errs, p.Endpoint.err)
	}

	if (p.Auth.Bearer == "") && (len(p.Auth.Basic) == 0) && !p.Auth.Anonymous {
		errs = append(errs, stringError("'auth' must have either 'anon' or 'bearer/basic'"))
	}

	if p.ID == "" {
		errs = append(errs, requiredError(ID))
	}
	if p.Topic == "" {
		errs = append(errs, requiredError(TOPIC))
	}

	return errs, true
}

// Connect to a Kafka server
func (p *kafkaProducer) Connect() (Producer, error) {

	log.Info().Msgf("%s: connecting...", p.ID)

	cfg := sarama.NewConfig()
	cfg.ClientID = fmt.Sprintf("melp-sender-%s", p.ID)
	cfg.Producer.RequiredAcks = sarama.WaitForLocal
	cfg.Producer.Retry.Max = 10
	cfg.Producer.Return.Successes = true
	cfg.Producer.Compression = sarama.CompressionSnappy
	//cfg.Producer.Flush.Frequency = 500 * time.Millisecond

	p.Endpoint.SetConfig(cfg)

	var list = p.Endpoint.Peers()

	producer, err := sarama.NewSyncProducer(list, cfg)
	if err != nil {
		return nil, err
	}
	p.msg = producer

	return p, nil
}

func (p *kafkaProducer) Close() error {
	if !p.connected {
		return nil
	}
	err := p.msg.Close()
	p.connected = false
	return err
}

func (p *kafkaProducer) Authorize(r *http.Request) (bool, error) {
	if p.Auth == nil {
		return true, nil
	}
	return p.Auth.Validate(r)
}

type kafkaSendResponse struct {
	Partition int32 `json:"partition"`
	Offset    int64 `json:"offset"`
}

func (p *kafkaProducer) Send(msg Message) (interface{}, error) {
	var hdrs []sarama.RecordHeader

	for k, v := range msg.Headers {
		hdrs = append(hdrs, sarama.RecordHeader{
			Key:   []byte(k),
			Value: []byte(v),
		})
	}

	var pkg = sarama.ProducerMessage{
		Topic:     p.Topic,
		Headers:   hdrs,
		Timestamp: time.Now().UTC(),
		Value:     sarama.ByteEncoder(msg.Body),
	}

	partition, offset, err := p.msg.SendMessage(&pkg)

	if err != nil {
		return nil, err
	}
	log.Trace().Msgf("%s: msg sent: %d/%d", p.ID, partition, offset)

	return &kafkaSendResponse{
		Partition: partition,
		Offset:    offset,
	}, nil
}
