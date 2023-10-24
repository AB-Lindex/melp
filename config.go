package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/ninlil/butler/log"
	"github.com/ninlil/envsubst"
	yaml "gopkg.in/yaml.v3"
)

type melpConfig struct {
	APIVersion string        `json:"apiVersion" yaml:"apiVersion"`
	Producing  melpProducer  `json:"producers" yaml:"producers"`
	Consuming  melpConsumer  `json:"consumers" yaml:"consumers"`
	Endpoint   melpEndpoint  `json:"endpoints" yaml:"endpoints"`
	Metrics    metricsConfig `json:"metrics" yaml:"metrics"`

	outputs map[string]Producer

	inputs   []Consumer
	wgListen *sync.WaitGroup
}

type metricsConfig struct {
	Go      bool `json:"go" yaml:"go"`
	Process bool `json:"process" yaml:"process"`
}

type melpProducer struct {
	Kafka []*melpKafkaOutputConfig `json:"kafka" yaml:"kafka"`
}

type melpConsumer struct {
	Kafka []*melpKafkaInputConfig `json:"kafka" yaml:"kafka"`
}

type melpEndpoint struct {
	Kafka []*melpKafkaEndpointConfig `json:"kafka" yaml:"kafka"`
}

var config = new(melpConfig)

func (cfg *melpConfig) Echo() {
	buf, err := yaml.Marshal(config)
	if err != nil {
		log.Fatal().Msgf("unable to marshal config: %v", err)
	}
	fmt.Println(string(buf))
}

var notFound = make(map[string]int)

func expandEnv(key string) (string, bool) {
	if str, ok := os.LookupEnv(key); ok {
		return str, true
	}
	notFound[key]++
	if settings.Echo != nil {
		return fmt.Sprintf("${%s !!NOT_FOUND!!}", key), true
	}
	return "", true
}

func (cfg *melpConfig) loadFromFile(name string) bool {
	f, err := os.Open(name)
	if err != nil {
		log.Error().Msgf("unable to load config: %v", err)
		return false
	}
	defer f.Close()

	var buf bytes.Buffer

	envsubst.SetPrefix('$')
	envsubst.SetWrapper('{')
	err = envsubst.Convert(f, &buf, expandEnv)
	if err != nil {
		log.Error().Msgf("unable to load config: %v", err)
		return false
	}

	if settings.Echo == nil {
		for nf := range notFound {
			log.Error().Msgf("unable to expand '%s'", nf)
		}
		if len(notFound) > 0 {
			return false
		}
	}

	parser := yaml.NewDecoder(&buf)
	parser.KnownFields(!settings.Relaxed)
	err = parser.Decode(cfg)
	if err != nil {
		log.Error().Msgf("unable to parse config: %v", err)
		return false
	}

	return true
}

func (cfg *melpConfig) Load(name string) bool {
	var inUse int

	outputMap := make(map[string]int)

	if !cfg.loadFromFile(name) {
		return false
	}

	var ok = true

	log.Trace().Msgf("validating '%s'...", name)
	for i, output := range cfg.Producing.Kafka {
		errs, active := output.Validate()
		if printErrors(errs, "Error validating output #%d:", i) {
			ok = false
		} else {
			if active {
				inUse++
				outputMap[output.ID]++
			}
		}
	}

	for i, input := range cfg.Consuming.Kafka {
		errs, active := input.Validate()
		if active {
			if printErrors(errs, "Error validating input #%d:", i) {
				ok = false
			} else {
				inUse++
			}
		}
	}

	for id, count := range outputMap {
		if count > 1 {
			log.Error().Msgf("error: duplicate output id '%s'", id)
			ok = false
		}
	}

	if inUse == 0 {
		log.Info().Msg("Nothing to do. Exiting")
		return settings.DryRun
	}

	return ok
}

func (cfg *melpConfig) Connect() bool {
	var ok = true

	if cfg.outputs == nil {
		cfg.outputs = make(map[string]Producer)
	}

	for _, output := range cfg.Producing.Kafka {
		kp, err := output.NewProducer()
		if err != nil {
			ok = false
		}

		if kp != nil {
			cfg.outputs[kp.Name()] = kp
		}
	}

	for _, input := range cfg.Consuming.Kafka {
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
