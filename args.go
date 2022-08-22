package main

import (
	"time"

	arg "github.com/alexflint/go-arg"
	"github.com/ninlil/butler/log"
)

type melpArgs struct {
	Config         string        `arg:"-f,--file,env:CONFIG" default:"melp.yaml" help:"name of config-file"`
	Port           int           `arg:"-p,--port,env:HTTP_PORT" default:"10000" help:"http-port number"`
	ReconnectTimer time.Duration `arg:"--reconnect-delay" help:"delay when reconnecting after failure" default:"15s" placeholder:"DELAY"`
	LogLevel       int           `arg:"-l,--loglevel" help:"log-level" default:"5" placeholder:"LEVEL"`
	Relaxed        bool          `arg:"--relax" help:"relaxed parsing of config-file"`
}

func (melpArgs) Version() string {
	return versionFunc()
}

var settings melpArgs

func init() {
	arg.MustParse(&settings)

	if settings.LogLevel < 1 {
		settings.LogLevel = 1
	}
	if settings.LogLevel > 8 {
		settings.LogLevel = 8
	}

	log.WithLevel(log.Level(settings.LogLevel))

	if settings.ReconnectTimer < time.Second {
		settings.ReconnectTimer = time.Second
	}
	if settings.ReconnectTimer > time.Minute {
		settings.ReconnectTimer = time.Minute
	}
}
