package main

import (
	"context"
	"os"
	"sync"

	"github.com/ninlil/butler/log"
	yaml "gopkg.in/yaml.v3"
)

type melpConfig struct {
	APIVersion string     `json:"apiVersion" yaml:"apiVersion"`
	Output     melpOutput `json:"output" yaml:"output"`
	Input      melpInput  `json:"intut" yaml:"input"`

	outputs map[string]Producer

	inputs   []Consumer
	wgListen *sync.WaitGroup
}

type melpOutput struct {
	Kafka []*melpKafkaOutputConfig `json:"kafka" yaml:"kafka"`
}

type melpInput struct {
	Kafka []*melpKafkaInputConfig `json:"kafka" yaml:"kafka"`
}

const (
	defaultConfigName = "melp.yaml"
)

var config = new(melpConfig)

func (cfg *melpConfig) Load(name string) bool {
	if name == "" {
		name = defaultConfigName
	}

	f, err := os.Open(name)
	if err != nil {
		log.Error().Msgf("unable to load config: %v", err)
		return false
	}
	defer f.Close()

	parser := yaml.NewDecoder(f)
	parser.KnownFields(true)
	err = parser.Decode(cfg)
	if err != nil {
		log.Error().Msgf("unable to parse config: %v", err)
		return false
	}

	var ok = true

	log.Trace().Msg("validating...")
	for i, output := range cfg.Output.Kafka {
		errs := output.Validate()
		if printErrors(errs, "Error validating output #%d:", i) {
			ok = false
		}
	}

	for i, input := range cfg.Input.Kafka {
		errs := input.Validate()
		if printErrors(errs, "Error validating input #%d:", i) {
			ok = false
		}
	}

	return ok
}

func (cfg *melpConfig) Connect() bool {
	var ok = true

	if cfg.outputs == nil {
		cfg.outputs = make(map[string]Producer)
	}

	for _, output := range cfg.Output.Kafka {
		kp, err := output.NewProducer()
		if err != nil {
			ok = false
		}

		if kp != nil {
			cfg.outputs[kp.Name()] = kp
		}
	}

	for _, input := range cfg.Input.Kafka {
		kc, errs := input.NewReceiver()
		if errs != nil {
			ok = false
		}

		if kc != nil {
			cfg.inputs = append(cfg.inputs, kc)
		}
	}

	return ok
}

func (cfg *melpConfig) Listen() {
	cfg.wgListen = &sync.WaitGroup{}
	for _, r := range cfg.inputs {
		log.Info().Msgf("starting to listen on '%s'...", r.Name())
		cfg.wgListen.Add(1)
		r.Listen(cfg.wgListen)
	}
}

func (cfg *melpConfig) Close(ctx context.Context) {
	log.Info().Msgf("Closing all connections...")
	for _, output := range config.outputs {
		err := output.Close()
		if err != nil {
			log.Warn().Msgf("error closing '%s': %v", output.Name(), err)
		}
	}

	for _, input := range config.inputs {
		err := input.Close()
		if err != nil {
			log.Warn().Msgf("error closing '%s': %v", input.Name(), err)
		}
	}

	if cfg.wgListen != nil {
		log.Debug().Msg("waiting for all listeners to close...")
		cfg.wgListen.Wait()
	}

	log.Info().Msg("all closed.")
}
