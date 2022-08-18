package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"time"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog/log"
)

var netClient = retryablehttp.NewClient()

func init() {
	netClient.RetryMax = 5
	netClient.RetryWaitMin = time.Millisecond * 500
	netClient.RetryWaitMax = time.Second * 3
}

type httpLogger struct{}

func (hlog *httpLogger) Format(format string, v ...interface{}) string {
	return fmt.Sprintf("http-client: "+format, v...)
}
func (hlog *httpLogger) Error(format string, v ...interface{}) {
	log.Error().Msg(hlog.Format(format, v...))
}
func (hlog *httpLogger) Info(format string, v ...interface{}) {
	log.Info().Msgf(hlog.Format(format, v...))
}
func (hlog *httpLogger) Debug(format string, v ...interface{}) {
	log.Debug().Msgf(hlog.Format(format, v...))
}
func (hlog *httpLogger) Warn(format string, v ...interface{}) {
	log.Warn().Msgf(hlog.Format(format, v...))
}

func (callback *melpCallback) Send(message *Message) error {

	netClient.Logger = new(httpLogger)

	log.Trace().Msgf("Send-> preparing to send message...")

	url := os.Expand(callback.URL, func(key string) string {
		return message.Metadata[key]
	})

	log.Trace().Msgf("Send-> url = '%s'", url)

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

	for k, v := range message.Headers {
		req.Header.Add(k, v)
	}

	resp, err := netClient.Do(req)

	if err != nil {
		log.Error().Msgf("send failed: %v", err)
		return err
	}

	if resp != nil {
		log.Debug().Msgf("Send-> send-response = %d", resp.StatusCode)
	}

	return err
}
