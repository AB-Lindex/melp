package main

import (
	"fmt"

	"github.com/ninlil/butler/log"
)

type melpKafkaEndpointConfig struct {
	Name     string `json:"name" yaml:"name"`
	Endpoint string `json:"endpoint" yaml:"endpoint"`
	Key      string `json:"key" yaml:"key"`
	Secret   string `json:"secret" yaml:"secret"`
}

type melpKafkaOutputConfig struct {
	Endpoint string `json:"endpoint" yaml:"endpoint"`
	ID       string `json:"id" yaml:"id"`
	Disabled bool   `json:"disabled" yaml:"disabled"`

	Topic string `json:"topic" yaml:"topic"`

	Auth Auth `json:"auth" yaml:"auth"`

	producer *kafkaProducer
}

type melpKafkaInputConfig struct {
	Endpoint string `json:"endpoint" yaml:"endpoint"`
	ID       string `json:"id" yaml:"id"`
	Disabled bool   `json:"disabled" yaml:"disabled"`

	Topics []string `json:"topics" yaml:"topics"`
	Group  string   `json:"group" yaml:"group"`

	Callback melpCallback `json:"callback" yaml:"callback"`

	consumer *kafkaReceiver
}

type melpCallback struct {
	URL     string            `json:"url" yaml:"url"`
	Auth    *Auth             `json:"auth" yaml:"auth"`
	Headers map[string]string `json:"headers" yaml:"headers"`
}

func (config *melpKafkaOutputConfig) Validate() ([]error, bool) {

	if config.Disabled {
		return nil, false
	}

	ep := findKafkaEndpoint(config.Endpoint)
	if ep == nil {
		return []error{stringError(fmt.Sprintf("endpoint '%s' not found", config.Endpoint))}, false
	}

	config.producer = &kafkaProducer{
		ID:       config.ID,
		Endpoint: newKafkaEndpoint(ep),
		Topic:    config.Topic,
		Auth:     &config.Auth,
	}

	return config.producer.Validate()
}

func (config *melpKafkaOutputConfig) NewProducer() (Producer, error) {
	if config.Disabled {
		return nil, nil
	}

	if config.producer == nil {
		log.Panic().Msg("config.producer is null")
	}

	log.Info().Msgf("Connecting to '%s'...", config.producer.Name())
	_, err := config.producer.Connect()
	if err != nil {
		log.Error().Msgf("%s: unable to connect: %v", config.producer.Name(), err)
		return nil, err
	}
	return config.producer, nil
}

func (config *melpKafkaInputConfig) Validate() ([]error, bool) {
	if config.Disabled {
		return nil, false
	}

	ep := findKafkaEndpoint(config.Endpoint)
	if ep == nil {
		return []error{stringError(fmt.Sprintf("endpoint '%s' not found", config.Endpoint))}, false
	}

	config.consumer = &kafkaReceiver{
		ID:       config.ID,
		Endpoint: newKafkaEndpoint(ep),
		Topics:   config.Topics,
		Group:    config.Group,
		Callback: config.Callback,
	}

	return config.consumer.Validate()
}

func (config *melpKafkaInputConfig) NewReceiver() (Consumer, error) {
	if config.Disabled {
		return nil, nil
	}

	log.Info().Msgf("Listening to '%s'...", config.consumer.Name())
	_, err := config.consumer.Connect()
	if err != nil {
		log.Error().Msgf("%s: unable to listen to: %v", config.consumer.Name(), err)
		return nil, nil
	}

	return config.consumer, nil
}
