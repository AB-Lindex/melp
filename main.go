package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/ninlil/butler"
	"github.com/ninlil/butler/log"
	"github.com/ninlil/butler/router"
)

var routes = []router.Route{
	{Name: "dump", Method: "POST", Path: "/dump", Handler: dump},
	{Name: "fail", Method: "POST", Path: "/fail", Handler: fail},
	{Name: "send", Method: "POST", Path: "/send/{id}", Handler: send},
}

func main() {
	os.Exit(func() int {
		defer butler.Cleanup(config.Close)

		err := router.Serve(routes, router.WithPort(9090))
		if err != nil {
			log.Fatal().Msg(err.Error())
		}

		if !config.Load("") {
			return 1
		}

		if !config.Connect() {
			log.Warn().Msg("some integrations failed")
		}

		config.Listen()

		butler.Run()
		return 0
	}())
}

func dump(r *http.Request) {

	fmt.Printf("Dump URL: %s\n", r.URL)

	var params map[string][]string = r.URL.Query()
	if len(params) > 0 {
		fmt.Println("Query params:")
		for k, v := range params {
			fmt.Printf(" - %s: %s\n", k, strings.Join(v, ", "))
		}
	}

	fmt.Println("Headers:")
	for k, v := range r.Header {
		fmt.Printf(" - %s: %s\n", k, strings.Join(v, ";"))
	}
	fmt.Println("Body:")
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		log.Error().Msgf("error reading body: %v", err)
	} else {
		lines := strings.Split(string(body), "\n")
		for _, line := range lines {
			fmt.Printf(" > %s\n", line)
		}
	}
	log.Info().Msg("dump complete")
}

type failArgs struct {
	Status int `json:"status" from:"query"`
}

func fail(args *failArgs) int {
	return args.Status
}
