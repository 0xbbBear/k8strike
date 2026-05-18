package util

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"strings"

	"k8strike/pkg/errors"
)

func UnixHttpSend(method string, unixPath string, uri string, data string) (string, error) {
	httpc := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", unixPath)
			},
		},
	}

	var response *http.Response
	var err error

	switch method {
	case "post":
		response, err = httpc.Post(uri, "application/json", strings.NewReader(data))
	case "get":
		response, err = httpc.Get(uri)
	}

	if err != nil {
		return "", &errors.K8strikeRuntimeError{Err: err, CustomMsg: "Unix HTTP Request failed."}
	}
	defer response.Body.Close()
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, response.Body); err != nil {
		return "", &errors.K8strikeRuntimeError{Err: err, CustomMsg: "Failed to read response body"}
	}
	return buf.String(), nil
}

func HttpSendJson(method string, url string, data string) (string, error) {
	req, err := http.NewRequest(strings.ToUpper(method), url, bytes.NewBuffer([]byte(data)))
	if err != nil {
		return "", &errors.K8strikeRuntimeError{Err: err, CustomMsg: "HTTP Request failed."}
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", &errors.K8strikeRuntimeError{Err: err, CustomMsg: "HTTP Request failed."}
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", &errors.K8strikeRuntimeError{Err: err, CustomMsg: "Failed to read HTTP response body"}
	}
	return string(body), nil
}
