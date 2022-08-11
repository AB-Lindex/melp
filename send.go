package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/ninlil/butler/log"
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

	log.Info().Msgf("Send to '%s' (%s):", args.ID, p.Name())

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Msgf("error reading body: %v", err)
	}
	defer r.Body.Close()

	fmt.Println(string(body))

	msg := Message{Body: body}

	for _, h := range passthruHeaders {
		msg.AddHeader(h, r.Header.Get(h))
	}

	response, err := p.Send(msg)
	if err != nil {
		log.Error().Msgf("error sending msg: %v", err)
	}

	return response, http.StatusOK, nil
}
