package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)


func createRequest(method string, url string, body any) (*http.Request, error) {

	var reader io.Reader = nil;

	switch body := body.(type) {
		case nil:
			reader = nil;
		case string:
			reader = strings.NewReader(body);
		case []byte:
			reader = bytes.NewReader(body);
		case io.Reader:
			reader = body;
		default:
			jdata, err := json.Marshal(body);
			if err != nil {
				return nil, fmt.Errorf("failed to encode json: %m", err);
			}
			reader = bytes.NewReader(jdata);
	}

	req, err := http.NewRequest(method, url, reader);
	if err != nil {
		return nil, err;
	}

	// impersonate the offical hytale launcher ..
	// for some reason this makes our refresh token actually stick :?

	req.Header.Set("user-agent", "hytale-launcher/"+wCommune.SpoofLauncherVersion );
	req.Header.Set("x-hytale-launcher-version", wCommune.SpoofLauncherVersion );
	req.Header.Set("x-hytale-launcher-branch", "release");


	return req, nil;
}
