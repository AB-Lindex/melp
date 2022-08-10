package main

import (
	"encoding/base64"
	"net/http"
	"os"
	"strings"
)

// Auth is in what ways we protect the REST endpoints, both in and out
type Auth struct {
	Fail      bool              `json:"fail" yaml:"fail"`
	Anonymous bool              `json:"anon" yaml:"anon"`
	Bearer    string            `json:"bearer" yaml:"bearer"`
	Basic     map[string]string `json:"basic" yaml:"basic"`
}

var (
	//errAnonymousNotEnabled = stringError("Anonymous is not enabled")
	errAuthRequired    = stringError("Authorization-header is missing")
	errAuthMalformed   = stringError("Authorization-header is mal-formed")
	invalidAuthBearer  = invalidError("Auth-Bearer")
	invalidAuthBasic   = invalidError("Auth-Basic")
	invalidAuthUnknown = stringError("unknown user")
	errFail            = stringError("forced-fail")
)

// ExpandEnv uses `os.ExpandEnv` on all values (and usernames in 'Basic')
func (auth *Auth) ExpandEnv() *Auth {
	if auth == nil {
		return nil
	}
	auth.Bearer = os.ExpandEnv(auth.Bearer)

	if len(auth.Basic) > 0 {
		var basics = make(map[string]string)
		for u, p := range auth.Basic {
			u = os.ExpandEnv(u)
			p = os.ExpandEnv(p)
			if u != "" && p != "" { // dont add empty usernames or passwords if they became blank ENV-values
				basics[u] = p
			}
		}
		auth.Basic = basics
	}

	return auth
}

// Validate that a request is authorized to pass
func (auth Auth) Validate(r *http.Request) (bool, error) {

	if err := auth.VerifyAuthorization(r); err != nil {
		return false, err
	}

	if auth.Fail {
		return false, errFail
	}
	return true, nil
}

// VerifyAuthorization checks the 'Authorization' header, if needed
func (auth Auth) VerifyAuthorization(r *http.Request) error {
	//	if auth.Bearer == "" && len(auth.Basic) == 0 {
	if auth.Anonymous {
		return nil
	}
	// 	return errAnonymousNotEnabled
	// }

	header := r.Header.Get("Authorization")
	if header == "" {
		return errAuthRequired
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 {
		return errAuthMalformed
	}

	switch strings.ToLower(parts[0]) {
	case "basic":
		decoded, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			return invalidAuthBasic
		}

		creds := strings.SplitN(string(decoded), ":", 2)
		if len(creds) == 2 {
			if auth.Basic[creds[0]] == creds[1] {
				return nil
			}
			return invalidAuthUnknown
		}
		return invalidAuthBasic

	case "bearer":
		if parts[1] == auth.Bearer {
			return nil
		}
		return invalidAuthBearer
	}
	return errAuthMalformed
}
