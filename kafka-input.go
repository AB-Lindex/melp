package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/ninlil/butler/log"
)

type kafkaReceiver struct {
	client    sarama.ConsumerGroup
	connected bool

	ID       string
	Endpoint *kafkaEndpoint
	Topics   []string
	Group    string
	Callback melpCallback

	ready  chan bool
	ctx    context.Context
	cancel func()
	wg     *sync.WaitGroup
}

// type receiverCallback struct {
// 	URL  string
// 	Auth *Auth
// }

// // NewKafkaConsumer creates a Kafka consumer
// func NewKafkaConsumer(id, endpoint, key, secret, group string, topics []string) Consumer {
// 	return &kafkaReceiver{
// 		ID:       id,
// 		Endpoint: newKafkaEndpoint(endpoint, key, secret),
// 		Topics:   topics,
// 		Group:    group,
// 	}
// }

func (r *kafkaReceiver) Name() string {
	return r.ID
}

func (r *kafkaReceiver) Validate() []error {
	var errs []error

	if r.Callback.Auth != nil {
		if r.Callback.Auth.Bearer != "" && len(r.Callback.Auth.Basic) > 0 {
			errs = append(errs, fmt.Errorf("can't use both 'bearer' and 'basic' auth"))
		}
		r.Callback.Auth.Anonymous = (r.Callback.Auth.Bearer == "") && (len(r.Callback.Auth.Basic) == 0)
	}

	if r.Endpoint.err != nil {
		errs = append(errs, r.Endpoint.err)
	}

	if len(r.Topics) == 0 {
		errs = append(errs, requiredError(TOPICS))
	}

	if r.ID == "" {
		errs = append(errs, requiredError(ID))
	}

	if r.Group == "" {
		errs = append(errs, requiredError(GROUP))
	}

	if r.Callback.URL == "" {
		errs = append(errs, requiredError(URL))
	}

	return errs
}

func (r *kafkaReceiver) Close() error {
	if !r.connected {
		return nil
	}
	log.Trace().Msgf("Closing listener '%s'...", r.ID)
	r.cancel()
	r.connected = false
	return nil
}

// Setup is run at the beginning of a new session, before ConsumeClaim.
func (r *kafkaReceiver) Setup(session sarama.ConsumerGroupSession) error {
	log.Trace().Str("group", r.Group).Msgf("Kafka.Setup(%s) #%d", r.ID, session.GenerationID())
	// Mark the consumer as ready
	close(r.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
// but before the offsets are committed for the very last time.
func (r *kafkaReceiver) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Trace().Str("group", r.Group).Msgf("Kafka.Cleanup(%s) #%d", r.ID, session.GenerationID())
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
// Once the Messages() channel is closed, the Handler must finish its processing
// loop and exit.
func (r *kafkaReceiver) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	log.Trace().Str("group", r.Group).Msgf("Kafka.ConsumeClaim(%s) #%d", r.ID, session.GenerationID())
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/main/consumer_group.go#L27-L29
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				log.Warn().Msgf("%s: received -nil- message", r.ID)
			} else {
				log.Trace().Str("group", r.Group).Str("topic", message.Topic).Msgf("%s: Message claimed: value = %s, timestamp = %v", r.ID, string(message.Value), message.Timestamp)

				msg := r.CreateMessage(message)
				err := r.Callback.Send(msg)
				log.Trace().Msgf("r.Callback.Send(msg) -> %v", err)
				if err != nil {
					log.Error().
						Str("topic", message.Topic).
						Int32("partition", message.Partition).
						Int64("offset", message.Offset).
						Msgf("processing failed: %v", err)
					r.Reconnect(message.Topic, message.Partition, message.Offset)
				} else {
					session.MarkMessage(message, "")
				}
			}

		// Should return when `session.Context()` is done.
		// If not, will raise `ErrRebalanceInProgress` or `read tcp <ip>:<port>: i/o timeout` when kafka rebalance. see:
		// https://github.com/Shopify/sarama/issues/1192
		case <-session.Context().Done():
			return nil
		}
	}
}

func (r *kafkaReceiver) CreateMessage(message *sarama.ConsumerMessage) *Message {
	var msg = &Message{
		Body:      message.Value,
		Timestamp: message.Timestamp,
	}
	msg.AddMetadata("topic", message.Topic)
	msg.AddMetadata("partition", strconv.FormatInt(int64(message.Partition), 10))
	msg.AddMetadata("offset", strconv.FormatInt(message.Offset, 10))

	for _, h := range message.Headers {
		msg.AddHeader(string(h.Key), string(h.Value))
	}

	return msg
}

func (r *kafkaReceiver) Connect() (Consumer, error) {

	//sarama.Logger = stdlog.New(os.Stdout, "sarama", 0)

	cfg := sarama.NewConfig()
	cfg.ClientID = fmt.Sprintf("melp-reader-%s", r.ID)

	//cfg.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategySticky
	cfg.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin

	r.Endpoint.SetConfig(cfg)

	client, err := sarama.NewConsumerGroup(r.Endpoint.Peers(), r.Group, cfg)
	if err != nil {
		return nil, err
	}

	r.client = client

	return r, nil
}

func (r *kafkaReceiver) Reconnect(topic string, partition int32, offset int64) {
	log.Warn().Msgf("Kafka(%s): Attempting reconnect..", r.ID)
	go func() {
		r.client.Close()
		log.Trace().Msgf("RETRY - Closed... sleeping %s...", settings.ReconnectTimer)

		time.Sleep(settings.ReconnectTimer)

		log.Trace().Msg("RETRY - Re-connecting...")
		_, err := r.Connect()
		if err != nil {
			log.Panic().Msgf("Kafka(%s): reconnection failed: %v", r.ID, err)
		}
		r.wg.Add(1) // restart the listener (and waitgroup)
		r.Listen(r.wg)
	}()
}

func (r *kafkaReceiver) Listen(wg *sync.WaitGroup) {
	ctx, cancel := context.WithCancel(context.Background())
	r.wg = wg
	r.ctx = ctx
	r.cancel = cancel
	r.ready = make(chan bool)

	go func() {
		defer wg.Done()
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := r.client.Consume(ctx, r.Topics, r); err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					log.Warn().Msg("Kafka-Listen-error: Consumer-Group was closed")
					return
				}
				log.Panic().Msgf("Error from consumer: %v", err)
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				log.Info().Msgf("Kafka-Listen(%s) - closing down", r.ID)
				return
			}
			r.ready = make(chan bool)
		}
	}()

	<-r.ready // Await till the consumer has been set up
	r.connected = true
	log.Info().Msgf("Listener '%s' up and running...", r.ID)
}
