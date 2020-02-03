package main

import (
	"crypto/tls"
	"net/http"
)

func getHTTPClient(allowInsecure bool) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: allowInsecure,
			},
		},
	}
}
