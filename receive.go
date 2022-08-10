package main

import (
	"encoding/base64"
	"fmt"
	"os"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog/log"
)

var netClient = retryablehttp.NewClient()

//&http.Client{
//	Timeout: time.Second * 10,
//}

func (callback *melpCallback) Send(message *Message) error {

	log.Trace().Msgf("preparing to send message...")

	url := os.Expand(callback.URL, func(key string) string {
		return message.Metadata[key]
	})

	log.Logger.Trace().Msgf("  -> url = '%s'", url)

	//buffer := bytes.NewBuffer(message.Body)

	//resp, err := netClient.Post(url, message.ContentType(), buffer)

	req, err := retryablehttp.NewRequest("POST", url, message.Body)
	if err != nil {
		log.Error().Msgf("create request failed: %v", err)
		return err
	}

	if callback.Auth != nil {
		if callback.Auth.Bearer != "" {
			req.Header.Add("Autorization", fmt.Sprintf("Bearer %s", callback.Auth.Bearer))
		}
		if len(callback.Auth.Basic) > 0 {
			var text string
			for k, v := range callback.Auth.Basic {
				text = fmt.Sprintf("%s:%s", k, v)
			}
			basic := base64.StdEncoding.EncodeToString([]byte(text))
			req.Header.Add("Autorization", fmt.Sprintf("Basic %s", basic))
		}
	}

	for k, v := range message.Metadata {
		req.Header.Add(fmt.Sprintf("melp-%s", k), v)
	}

	resp, err := netClient.Do(req)

	if err != nil {
		log.Error().Msgf("send failed: %v", err)
		return err
	}

	if resp != nil {
		log.Debug().Msgf("send-response = %d", resp.StatusCode)
	}

	return err
}
