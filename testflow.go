//go:build testflow
// +build testflow

package main

// This file is meant to serve as testing features, providing as different "input callbacks"
// * dumping messages
// * intentional fail
// * retry (a.k.a resend), to speed-test an infinity-loop

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ninlil/butler/log"
	"github.com/ninlil/butler/router"
)

func init() {
	routes = append(routes,
		router.Route{Name: "dump", Method: "POST", Path: "/dump", Handler: dump},
		router.Route{Name: "fail", Method: "POST", Path: "/fail", Handler: fail},
		router.Route{Name: "retry", Method: "POST", Path: "/retry/{id}", Handler: retry})
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
	Status int    `json:"status" from:"query"`
	Accept string `json:"accept" from:"query"`
	Body   []byte `json:"body" from:"body"`
}

func fail(args *failArgs) int {
	if args.Accept != "" && strings.Contains(string(args.Body), args.Accept) {
		return 200
	}
	return args.Status
}

type retryArgs struct {
	ID   string `json:"id" from:"path"`
	Body string `from:"body"`
}

var stats struct {
	start time.Time
	count int
	ts    time.Time
}

func retry(args *retryArgs) {
	i, err := strconv.Atoi(string(args.Body))
	if err != nil {
		log.Error().Msgf("retry(%s) error: %v", args.Body, err)
		return
	}

	if output, ok := config.outputs[args.ID]; ok {
		if _, err := output.Send(Message{Body: []byte(fmt.Sprintf("%d", i+1))}); err != nil {
			log.Error().Msgf("retry(%d) failed: %v", i+1, err)
			return
		}

		if stats.count == 0 {
			stats.start = time.Now()
			stats.ts = stats.start
		}
		stats.count++

		if stats.ts.Add(time.Second).Before(time.Now()) {
			stats.ts = time.Now()
			dur := time.Since(stats.start)
			secs := dur.Seconds()
			var speed float64
			if secs > 0 {
				speed = float64(stats.count) / secs
			}
			log.Info().Msgf("##### Retry-stats: %d msg in %v -> %g msg/s", stats.count, dur, speed)
		}
	}
}
