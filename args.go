package main

import (
	"time"

	arg "github.com/alexflint/go-arg"
	"github.com/ninlil/butler/log"
)

type melpArgs struct {
	Config          string        `arg:"-f,--file,env:CONFIG" default:"melp.yaml" help:"name of config-file"`
	Port            int           `arg:"-p,--port,env:HTTP_PORT" default:"10000" help:"http-port number"`
	ReconnectDelay  time.Duration `arg:"--reconnect-delay,env:RECONNECT_DELAY" help:"delay when reconnecting after failure" default:"10s" placeholder:"DELAY"`
	ReconnectJitter time.Duration `arg:"--reconnect-jitter,env:RECONNECT_JITTER" help:"jitter when reconnecting after failure" default:"2s" placeholder:"JITTER"`
	LogLevel        int           `arg:"-l,--loglevel,env:LOGLEVEL" help:"log-level" default:"5" placeholder:"LEVEL"`
	Relaxed         bool          `arg:"--relax,env:CONFIG_RELAX" help:"relaxed parsing of config-file"`
	AllowStop       bool          `arg:"--allow-stop,env:ALLOW_STOP" help:"allowed stop running melp"`
	DryRun          bool          `arg:"--dry-run" help:"dry-run mode"`
	Echo            *echoCmd      `arg:"subcommand:echo" help:"print parsed config"`
}

type echoCmd struct{}

func (melpArgs) Version() string {
	return versionFunc()
}

var settings melpArgs

const (
	minReconnect = time.Second
	maxReconnect = time.Minute * 2
)

func init() {
	arg.MustParse(&settings)

	if settings.Echo != nil {
		settings.DryRun = true
		settings.Relaxed = true
	}

	if settings.LogLevel < 1 {
		settings.LogLevel = 1
	}
	if settings.LogLevel > 8 {
		settings.LogLevel = 8
	}

	log.WithLevel(log.Level(settings.LogLevel))

	if settings.ReconnectDelay < minReconnect {
		settings.ReconnectDelay = minReconnect
	}
	if settings.ReconnectDelay > maxReconnect {
		settings.ReconnectDelay = maxReconnect
	}

	settings.ReconnectJitter = settings.ReconnectJitter.Abs()
	if settings.ReconnectJitter > settings.ReconnectDelay {
		settings.ReconnectJitter = settings.ReconnectDelay
	}
}
