package main

import (
	"net/http"
	"os"

	"github.com/ninlil/butler"
	"github.com/ninlil/butler/log"
	"github.com/ninlil/butler/router"
)

var routes = []router.Route{
	{Name: "send", Method: "POST", Path: "/send/{id}", Handler: send},
	{Name: "stop", Method: "GET", Path: "/stop", Handler: stop},
}

func main() {
	initKafka()

	os.Exit(run())
}

func stop() int {
	if !settings.AllowStop {
		return http.StatusForbidden
	}
	butler.Quit()
	return http.StatusOK
}

func run() int {
	if !settings.DryRun {
		defer butler.Cleanup(config.Close)

		routes = append(routes,
			router.Route{Name: "metrics", Method: "GET", Path: "/metrics", Handler: metrics.Init()},
		)

		err := router.Serve(routes, router.WithPort(settings.Port))
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	if !config.Load(settings.Config) {
		return 1 // dont want os.Exit here, because then the deferred cleanup wouldn't trigger
	}

	if settings.Echo != nil {
		config.Echo()
		return 0
	}

	if settings.DryRun {
		log.Info().Msg("dry-run mode")
		return 0
	}

	if !config.Connect() {
		log.Warn().Msg("some integrations failed")
	}

	config.Listen()

	log.Info().Msgf("Reconnection delay: %v +/- %v", settings.ReconnectDelay, settings.ReconnectJitter)
	log.Info().Msgf("Stop-command %s", flag2text(settings.AllowStop, "enabled", "disabled"))

	butler.Run()
	return 0
}
