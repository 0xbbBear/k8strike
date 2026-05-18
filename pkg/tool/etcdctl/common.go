package etcdctl

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"k8strike/pkg/errors"
	"github.com/tidwall/gjson"
)

type EtcdRequestOption struct {
	Endpoint  string
	Api       string
	PostData  string
	TlsConfig *tls.Config
	Method    string
	Silent    bool
}

func DoRequest(opt EtcdRequestOption) (string, error) {
	if opt.TlsConfig == nil || len(opt.TlsConfig.Certificates) == 0 || opt.TlsConfig.RootCAs == nil {
		opt.TlsConfig = &tls.Config{InsecureSkipVerify: true}
	}
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: opt.TlsConfig,
		},
		Timeout: time.Duration(5) * time.Second,
	}

	request, err := http.NewRequest(opt.Method, opt.Endpoint+opt.Api, bytes.NewBuffer([]byte(opt.PostData)))
	if err != nil {
		return "", &errors.K8strikeRuntimeError{Err: err, CustomMsg: "err found while generate post request in net.http ."}
	}
	request.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(request)
	if resp != nil {
		defer resp.Body.Close()
	} else if err != nil {
		return "", &errors.K8strikeRuntimeError{Err: err, CustomMsg: "err found in post request."}
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", &errors.K8strikeRuntimeError{Err: err, CustomMsg: "err found in post request."}
	}

	return string(content), nil
}

func GetKeys(content string, silent bool) (map[string]string, error) {
	kvs := gjson.Get(content, "kvs").Array()
	ret := make(map[string]string, len(kvs))
	for _, k := range kvs {
		name, err := base64.StdEncoding.DecodeString(k.Get("key").String())
		if err != nil {
			fmt.Println("base64 decode failed:", err.Error())
			continue
		}

		ret[string(name)] = ""
		if !silent {
			fmt.Println(string(name))
		}

		if k.Get("value").Exists() {
			v, _ := base64.StdEncoding.DecodeString(k.Get("value").String())
			if !silent {
				fmt.Println(string(v))
			}
			ret[string(name)] = string(v)
		}
	}
	return ret, nil
}

func GenerateQuery(key string) (query string) {
	b64key := base64.StdEncoding.EncodeToString([]byte(strings.TrimSuffix(key, "\n")))
	if key == "/" {
		bzero := base64.StdEncoding.EncodeToString([]byte{0})
		query = fmt.Sprintf("{\"range_end\": \"%s\", \"key\": \"%s\", \"keys_only\":true}", bzero, b64key)
	} else {
		query = fmt.Sprintf("{\"key\": \"%s\"}", b64key)
	}
	return
}

func GetVersion(endpoint string) (string, string, error) {
	opt := EtcdRequestOption{
		Endpoint: endpoint,
		Api:      "/version",
		Method:   "GET",
	}
	resp, err := DoRequest(opt)
	if err != nil {
		return "", "", err
	}
	sv := gjson.Get(resp, "etcdserver").String()
	cv := gjson.Get(resp, "etcdcluster").String()
	return sv, cv, nil
}
