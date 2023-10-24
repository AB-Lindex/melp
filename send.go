package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ninlil/butler/log"
	"github.com/ninlil/butler/router"
)

type sendArgs struct {
	ID string `from:"path" json:"id" required:""`
}

var (
	errNoSuchProducer = fmt.Errorf("no such output")
)

func send(args *sendArgs, r *http.Request) (interface{}, int, error) {
	p := config.outputs[args.ID]

	if p == nil {
		return nil, http.StatusBadRequest, errNoSuchProducer
	}

	if ok, err := p.Authorize(r); !ok {
		if err != nil {
			log.Warn().Msgf("Auth failure '%s': %v", p.Name(), err)
		}
		return nil, http.StatusUnauthorized, nil
	}

	log.Trace().Msgf("Send to '%s' (%s):", args.ID, p.Name())

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Msgf("error reading body: %v", err)
	}
	defer r.Body.Close()

	msg := Message{Body: body}

	msg.AddHeader("X-Request-Id", router.ReqIDFromRequest(r))
	msg.AddHeader("X-Correlation-Id", router.CorrIDFromRequest(r))

	for _, h := range passthruHeaders {
		msg.AddHeader(h, r.Header.Get(h))
	}
	hdrs := map[string][]string(r.Header)
	for key, vs := range hdrs {
		if len(vs) > 0 {
			key = strings.ToLower(key)
			if strings.HasPrefix(key, "melp-") {
				key = key[5:]
				msg.AddHeader(key, vs[len(vs)-1])
			}
		}
	}

	response, err := p.Send(msg, r)
	if err != nil {
		log.Error().Msgf("error sending msg: %v", err)
	}

	return response, http.StatusOK, nil
}
