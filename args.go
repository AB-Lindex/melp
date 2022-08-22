package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	arg "github.com/alexflint/go-arg"
	"github.com/ninlil/butler/log"
)

type melpArgs struct {
	Config         string        `arg:"-f,--file,env:CONFIG" default:"melp.yaml" help:"name of config-file"`
	Port           int           `arg:"--port,env:HTTP_PORT" default:"10000" help:"http-port number"`
	ReconnectTimer time.Duration `arg:"--reconnect-delay" help:"delay when reconnecting after failure" default:"15s" placeholder:"DELAY"`
	LogLevel       int           `arg:"-l,--loglevel" help:"log-level" default:"5" placeholder:"LEVEL"`
	Relaxed        bool          `arg:"--relax" help:"relaxed parsing of config-file"`
}

func (melpArgs) Version() string {
	return versionFunc()
}

var settings melpArgs

func init() {
	parser, err := arg.NewParser(arg.Config{IgnoreEnv: true}, &settings)
	if err != nil {
		fmt.Println("argument error:", err)
		os.Exit(1)
	}

	err = parser.Parse(os.Args[1:])
	if err != nil {
		if errors.Is(err, arg.ErrHelp) {
			parser.WriteHelp(os.Stdout)
			os.Exit(0)
		}
		if errors.Is(err, arg.ErrVersion) {
			fmt.Print(versionFunc())
			os.Exit(0)
		}
		fmt.Println("error parsing argument:", err)
		os.Exit(1)
	}

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
