package main

import (
	"os"

	"github.com/ninlil/butler/log"
)

type melpKafkaOutputConfig struct {
	Endpoint string `json:"endpoint" yaml:"endpoint"`
	Key      string `json:"key" yaml:"key"`
	Secret   string `json:"secret" yaml:"secret"`
	ID       string `json:"id" yaml:"id"`
	Disabled bool   `json:"disabled" yaml:"disabled"`

	Topic string `json:"topic" yaml:"topic"`

	Auth Auth `json:"auth" yaml:"auth"`

	producer *kafkaProducer
}

type melpKafkaInputConfig struct {
	Endpoint string `json:"endpoint" yaml:"endpoint"`
	Key      string `json:"key" yaml:"key"`
	Secret   string `json:"secret" yaml:"secret"`
	ID       string `json:"id" yaml:"id"`
	Disabled bool   `json:"disabled" yaml:"disabled"`

	Topics []string `json:"topics" yaml:"topics"`
	Group  string   `json:"group" yaml:"group"`

	Callback melpCallback `json:"callback" yaml:"callback"`

	consumer *kafkaReceiver
}

type melpCallback struct {
	URL  string `json:"url" yaml:"url"`
	Auth *Auth  `json:"auth" yaml:"auth"`
}

func (config *melpKafkaOutputConfig) Validate() []error {

	if config.Disabled {
		return nil
	}

	config.producer = &kafkaProducer{
		ID:       os.ExpandEnv(config.ID),
		Endpoint: newKafkaEndpoint(config.Endpoint, config.Key, config.Secret),
		Topic:    os.ExpandEnv(config.Topic),
		Auth:     config.Auth.ExpandEnv(),
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

func (config *melpKafkaInputConfig) Validate() []error {
	if config.Disabled {
		return nil
	}

	for i := range config.Topics {
		config.Topics[i] = os.ExpandEnv(config.Topics[i])
	}

	config.consumer = &kafkaReceiver{
		ID:       os.ExpandEnv(config.ID),
		Endpoint: newKafkaEndpoint(config.Endpoint, config.Key, config.Secret),
		Topics:   config.Topics,
		Group:    os.ExpandEnv(config.Group),
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
