package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"time"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/ninlil/envsubst"
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

func expandMap(txt string, vars map[string]string) string {
	envsubst.SetPrefix('%')
	str, err := envsubst.ConvertString(txt, func(key string) (string, bool) {
		v, ok := vars[key]
		return v, ok
	})
	if err != nil {
		log.Error().Msgf("expandMap-error: %v", err)
	}
	return str
}

var melpUserAgent = fmt.Sprintf("melp-%s", versionFunc())

func (callback *melpCallback) Send(message *Message) error {
	netClient.Logger = new(httpLogger)

	log.Trace().Msgf("Send-> preparing to send message to '%s'...", callback.URL)

	// target := os.Expand(callback.URL, func(key string) string {
	// 	return message.Metadata[key]
	// })
	target := expandMap(callback.URL, message.Metadata)
	if _, err := url.Parse(target); err != nil {
		return fmt.Errorf("invalid url: %s", target)
	}

	log.Trace().Msgf("Send-> url = '%s'", target)

	//buffer := bytes.NewBuffer(message.Body)

	//resp, err := netClient.Post(url, message.ContentType(), buffer)

	req, err := retryablehttp.NewRequest("POST", target, message.Body)
	if err != nil {
		log.Error().Msgf("create request failed: %v", err)
		return err
	}

	if callback.Auth != nil {
		if callback.Auth.Bearer != "" {
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", callback.Auth.Bearer))
		}
		if len(callback.Auth.Basic) > 0 {
			var text string
			for k, v := range callback.Auth.Basic {
				text = fmt.Sprintf("%s:%s", k, v)
			}
			basic := base64.StdEncoding.EncodeToString([]byte(text))
			req.Header.Add("Authorization", fmt.Sprintf("Basic %s", basic))
		}
	}

	req.Header.Add("User-Agent", melpUserAgent)

	for k, v := range callback.Headers {
		req.Header.Add(k, v)
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

	if err == nil && (resp.StatusCode < http.StatusOK || resp.StatusCode > 299) {
		return fmt.Errorf(resp.Status)
	}

	return err
}
