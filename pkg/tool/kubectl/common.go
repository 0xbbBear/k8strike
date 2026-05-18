package kubectl

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"k8strike/pkg/conf"
	"k8strike/pkg/errors"
	"k8strike/pkg/util"
)

var MaybeSuccessfulStatuscodeList = []int{
	100,
	101,
	102,
	103,

	200,
	201,
	202,
	203,
	204,
	205,
	206,
	207,
	208,
	226,
}

type K8sRequestOption struct {
	TokenPath string
	Token     string
	Server    string
	Api       string
	Method    string
	PostData  string
	Url       string
	Anonymous bool
}

func ApiServerAddr() (string, error) {
	protocol := ""
	host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
	if len(host) == 0 || len(port) == 0 {
		text := "err: cannot find kubernetes api host in ENV"
		return "", errors.New(text)
	}
	if port == "8080" || port == "8001" {
		protocol = "http://"
	} else {
		protocol = "https://"
	}
	return protocol + net.JoinHostPort(host, port), nil
}

func GetServiceAccountToken(tokenPath string) (string, error) {
	token, err := os.ReadFile(tokenPath)
	if err != nil {
		return "", err
	}
	return string(token), nil
}

func SecretToken(tokenPath string) (string, error) {
	var tokenErr error
	var token string

	if tokenPath != "" {
		token, tokenErr = GetServiceAccountToken(tokenPath)
	} else if token == "" {
		token, tokenErr = GetServiceAccountToken(conf.K8sSATokenDefaultPath)
	}
	if tokenErr != nil {
		return "", &errors.K8strikeRuntimeError{Err: tokenErr, CustomMsg: "load K8s service account token error."}
	}

	token = strings.TrimSpace(token)

	return token, nil
}

func ServerAccountRequest(opts K8sRequestOption) (string, error) {

	if opts.Anonymous {
		opts.Token = ""
	} else if token, err := SecretToken(opts.TokenPath); err != nil {
		return "", err
	} else {
		opts.Token = token
	}

	if len(opts.Url) == 0 {
		var server string
		var urlErr error
		if opts.Server == "" {
			server, urlErr = ApiServerAddr()
			opts.Url = server + opts.Api
		} else {
			opts.Url = opts.Server + opts.Api
			urlErr = nil
		}
		if urlErr != nil {
			return "", &errors.K8strikeRuntimeError{Err: urlErr, CustomMsg: "err found while searching local K8s apiserver addr."}
		}
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: time.Second * 3,
	}
	var request *http.Request
	opts.Method = strings.ToUpper(opts.Method)

	request, err := http.NewRequest(opts.Method, opts.Url, bytes.NewBuffer([]byte(opts.PostData)))
	if err != nil {
		return "", &errors.K8strikeRuntimeError{Err: err, CustomMsg: "err found while generate post request in net.http ."}
	}

	if opts.Method == "POST" {
		request.Header.Set("Content-Type", "application/json")
	}
	if len(opts.Token) > 0 {
		token := strings.TrimSpace(opts.Token)
		request.Header.Set("Authorization", "Bearer "+token)
	}

	if opts.Anonymous {
		ips := make([]string, 100)

		for i := 0; i < 100; i++ {
			ip := fmt.Sprintf("10.%d.%d.%d", rand.Intn(256), rand.Intn(256), rand.Intn(256))
			ips[i] = ip
		}

		randIpStr := strings.Join(ips, ",")
		request.Header.Set("X-Forwarded-For", randIpStr)
	}

	resp, err := client.Do(request)
	if err != nil {
		return "", &errors.K8strikeRuntimeError{Err: err, CustomMsg: "err found in post request."}
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", &errors.K8strikeRuntimeError{Err: err, CustomMsg: "err found in post request."}
	}

	res := string(content)

	if !util.IntContains(MaybeSuccessfulStatuscodeList, resp.StatusCode) {
		errMsg := fmt.Sprintf("err found in post request, error response code: %v.", resp.Status)
		return res, &errors.K8strikeRuntimeError{
			Err:       err,
			CustomMsg: errMsg,
		}
	}

	return res, nil
}

func GetServerVersion(serverAddr string) (string, error) {
	opts := K8sRequestOption{
		TokenPath: "",
		Server:    serverAddr,
		Api:       "/version",
		Method:    "GET",
		PostData:  "",
		Anonymous: true,
	}
	resp, err := ServerAccountRequest(opts)
	if err != nil {
		return "", err
	}
	versionPattern := regexp.MustCompile(`"gitVersion":.*?"(.*?)"`)
	results := versionPattern.FindStringSubmatch(resp)
	if len(results) != 2 {
		return "", errors.New("field gitVersion not found in response")
	}
	return results[1], nil
}
