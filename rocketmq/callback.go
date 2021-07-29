package rocketmq

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"github.com/motemen/go-loghttp"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type Callback struct {
	client *http.Client
}

const (
	defaultTimeout = 10 * time.Second // 超时时间.
)

func NewCallback() *Callback {
	defaultTransport := http.DefaultTransport
	defaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: false}
	transport := &loghttp.Transport{
		Transport: defaultTransport,
	}
	return &Callback{
		client: &http.Client{Timeout: defaultTimeout, Transport: transport},
	}
}

type CallbackResponse struct {
	Code    int         `json:"code"`
	Msg     string      `json:"msg"`
	Data    interface{} `json:"data"`
	TraceId string      `json:"traceId"`
}

func (c *Callback) call(ctx context.Context, path string, body interface{}, cookie string) (code int, err error) {
	resp := &CallbackResponse{}

	josnBody, err := json.Marshal(body)
	if err != nil {
		return http.StatusBadRequest, err
	}

	httpRequest, err := http.NewRequest(http.MethodPut, path, strings.NewReader(string(josnBody)))
	if err != nil {
		return http.StatusNotAcceptable, err
	}
	httpRequest.Header.Add("format", "json")
	httpRequest.Header.Add("Cookie", cookie)

	response, err := c.client.Do(httpRequest)
	if err != nil {
		return http.StatusNotAcceptable, err
	}
	defer response.Body.Close()

	resBody, err := ioutil.ReadAll(response.Body)
	err = json.Unmarshal(resBody,&resp)
	return resp.Code, errors.WithStack(err)
}
