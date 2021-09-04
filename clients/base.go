package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-cygnus/constants"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"gopkg.in/go-playground/validator.v9"
	"gopkg.in/yaml.v3"
)

var logger *logrus.Entry = logrus.New().WithField("logger", "captain-client")

type restEndpoint struct {
	Scheme   string `yaml:"scheme" validate:"required"`
	Hostname string `yaml:"hostname" validate:"required"`
}

type restEndpointWithToken struct {
	Scheme   string `yaml:"scheme" validate:"required"`
	Hostname string `yaml:"hostname" validate:"required"`
	Token    string `yaml:"token" validate:"required"`
}

type clientsConfig struct {
	ApolloConfig *restEndpointWithToken `yaml:"apollo"`
}

// remote service url
type baseConfig struct {
	Timeout    int
	RetryTimes int
	Trace      bool
	Headers    map[string]string
}

const DefaultReqTimeSecond = 60

var (
	sharedClient = &http.Client{
		Timeout: DefaultReqTimeSecond * time.Second,
	}

	// TODO: not used, remove or add into http client hook
	defaultHTTPConfig = baseConfig{
		Timeout:    DefaultReqTimeSecond,
		RetryTimes: 0,
		Trace:      false,
	}

	RestConfigs clientsConfig
)

func init() {
	// increase default http client connection pool
	defaultRoundTripper := http.DefaultTransport

	defaultTransportPointer, ok := defaultRoundTripper.(*http.Transport)
	if !ok {
		panic(fmt.Sprintf("defaultRoundTripper not an *http.Transport"))
	}
	// copy it
	// dereference it to get a copy of the struct that the pointer points to
	copyTransportPointer := defaultTransportPointer
	copyTransportPointer.MaxIdleConns = 100
	copyTransportPointer.MaxIdleConnsPerHost = 100

	sharedClient.Transport = copyTransportPointer
}

func Init() {
	clientsYmlFile := fmt.Sprintf("%s/clients.yml", constants.ConfigPath)

	clientsContent, readErr := ioutil.ReadFile(clientsYmlFile)
	if readErr != nil {
		panic("Read clients.yml error.")
	}

	if unmarshalErr := yaml.Unmarshal(clientsContent, &RestConfigs); unmarshalErr != nil {
		panic(fmt.Sprintf("Unmarshal restclient content error: %s", unmarshalErr))
	}

	validate := validator.New()
	if err := validate.Struct(&RestConfigs); err != nil {
		panic(fmt.Sprintf("invalid restclient config: %s", err))
	}
}

type baseRest struct {
	Scheme  string
	Host    string
	URIBase string
	Config  baseConfig
}

func (b *baseRest) Request(method string, url string, payload interface{}) (req *http.Request, err error) {
	if payload == nil {
		payload = make(map[string]interface{})
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return
	}

	if req, err = http.NewRequest(method, url, bytes.NewBuffer(jsonBody)); err != nil {
		return
	}

	// set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	for k, v := range b.Config.Headers {
		req.Header.Set(k, v)
	}

	return
}

func (b *baseRest) Do(req *http.Request) (rspData []byte, err error) {
	startTime := time.Now()
	reqBody, _ := ioutil.ReadAll(req.Body)
	req.Body = ioutil.NopCloser(bytes.NewBuffer(reqBody))

	l := logger.WithField("url", req.URL.String()).WithField("req_body", string(reqBody))

	defer func() {
		if err != nil {
			l = l.WithError(err)
		}
		l = l.WithField("cost", time.Now().Sub(startTime).Seconds())
		l.Debug(time.Now().Sub(startTime).String())
	}()

	// TODO: can use b.Config.Trace to trace finer http lifecycle
	var rsp *http.Response
	if rsp, err = sharedClient.Do(req); err != nil {
		return
	}

	defer func() {
		_ = rsp.Body.Close()
		l = l.WithField("http_code", rsp.StatusCode)
	}()

	if rspData, err = ioutil.ReadAll(rsp.Body); err != nil {
		return
	}

	if rsp.StatusCode > 299 || rsp.StatusCode < 200 {
		err = errors.New(
			fmt.Sprintf("Invalid http %d when rest %s: %s", rsp.StatusCode, req.URL.String(), string(rspData)))

		return
	}

	return
}

func (b *baseRest) JsonWithReq(req *http.Request, out interface{}) (err error) {
	rspData, err := b.Do(req)
	if err != nil {
		return
	}

	if out != nil {
		err = json.Unmarshal(rspData, out)
	}

	return
}

func (b *baseRest) Json(method string, subPath string, payload interface{}, out interface{}) (err error) {
	url := fmt.Sprintf("%s://%s/%s/%s", b.Scheme, b.Host, b.URIBase, subPath)

	req, err := b.Request(method, url, payload)
	if err != nil {
		return
	}

	return b.JsonWithReq(req, out)
}

func SetLogger(l *logrus.Entry) {
	logger = l
}
