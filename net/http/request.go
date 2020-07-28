package http

import (
	"context"
	"net/http"
	"time"
)

var PrometheusImpl prometheus

type prometheus interface {
	HttpClientWithLabelValues(domain, httpCode, protocol, method, tag string, starTime time.Time)
}

type mockPrometheusImpl struct {
}

func (m *mockPrometheusImpl) HttpClientWithLabelValues(domain, httpCode, protocol, method, tag string, starTime time.Time) {
}

func init() {
	PrometheusImpl = new(mockPrometheusImpl)
}

// Request .
type Request interface {
	Post() Request
	Put() Request
	Get() Request
	Delete() Request
	Head() Request
	SetJSONBody(obj interface{}) Request
	SetBody(byts []byte) Request
	ToJSON(obj interface{}) Response
	ToString() (string, Response)
	ToBytes() ([]byte, Response)
	ToXML(v interface{}) Response
	SetHeader(header http.Header) Request
	AddHeader(key, value string) Request
	SetParam(key string, value ...interface{}) Request
	URL() string
	SetURL(uri string) Request
	SetContext(context.Context) Request
	Singleflight(key ...interface{}) Request
	SetName(name string) Request
	GetName() string
}

// Response .
type Response struct {
	Error         error
	Header        http.Header
	ContentLength int64
	ContentType   string
	StatusCode    int
	HTTP11        bool
}

type singleflightData struct {
	Res  Response
	Body []byte
}
