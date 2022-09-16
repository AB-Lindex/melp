package main

import (
	"os"

	"github.com/ninlil/butler"
	"github.com/ninlil/butler/log"
	"github.com/ninlil/butler/router"
)

var routes = []router.Route{
	{Name: "send", Method: "POST", Path: "/send/{id}", Handler: send},
}

func main() {
	initKafka()

	os.Exit(run())
}

func run() int {
	defer butler.Cleanup(config.Close)

	err := router.Serve(routes, router.WithPort(settings.Port))
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	if !config.Load(settings.Config) {
		return 1 // dont want os.Exit here, because then the deferred cleanup wouldn't trigger
	}

	if !config.Connect() {
		log.Warn().Msg("some integrations failed")
	}

	config.Listen()

	butler.Run()
	return 0
}
