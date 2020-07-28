package http

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	neturl "net/url"
	"strconv"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/sync/singleflight"
)

var (
	h2cclient      *http.Client
	h2cclientGroup singleflight.Group
)

func init() {
	Newh2cClient(15 * time.Second)
}

// Newh2cClient .
func Newh2cClient(rwTimeout time.Duration) {
	tran := &http2.Transport{
		AllowHTTP: true,
		DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
			fun := timeoutDialer(6 * time.Second)
			return fun(network, addr)
		},
	}

	h2cclient = &http.Client{
		Transport: tran,
		Timeout:   rwTimeout,
	}
}

// timeoutDialer returns functions of connection dialer with timeout settings for http.Transport Dial field.
func timeoutDialer(cTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		return conn, err
	}
}

// NewH2CRequest .
func NewH2CRequest(url ...string) Request {
	result := new(H2CRequest)
	req := &http.Request{
		Header: make(http.Header),
	}
	result.resq = req
	result.params = make(neturl.Values)
	if len(url) > 0 {
		result.url = url[0]
	}
	return result
}

// H2CRequest .
type H2CRequest struct {
	resq            *http.Request
	resp            *http.Response
	reqe            error
	params          neturl.Values
	url             string
	ctx             context.Context
	singleflightKey string
	name            string
}

// SetURL .
func (hr *H2CRequest) SetURL(uri string) Request {
	hr.url = uri
	return hr
}

// SetContext .
func (hr *H2CRequest) SetContext(ctx context.Context) Request {
	hr.ctx = ctx
	return hr
}

// Post .
func (hr *H2CRequest) Post() Request {
	hr.resq.Method = "POST"
	return hr
}

// Put .
func (hr *H2CRequest) Put() Request {
	hr.resq.Method = "PUT"
	return hr
}

// Get .
func (hr *H2CRequest) Get() Request {
	hr.resq.Method = "GET"
	return hr
}

// Get .
func (hr *H2CRequest) Head() Request {
	hr.resq.Method = "HEAD"
	return hr
}

// // Delete .
func (hr *H2CRequest) Delete() Request {
	hr.resq.Method = "DELETE"
	return hr
}

// SetJSONBody .
func (hr *H2CRequest) SetJSONBody(obj interface{}) Request {
	byts, e := json.Marshal(obj)
	if e != nil {
		hr.reqe = e
		return hr
	}

	hr.resq.Body = ioutil.NopCloser(bytes.NewReader(byts))
	hr.resq.ContentLength = int64(len(byts))
	hr.resq.Header.Set("Content-Type", "application/json")
	return hr
}

// SetBody .
func (hr *H2CRequest) SetBody(byts []byte) Request {
	hr.resq.Body = ioutil.NopCloser(bytes.NewReader(byts))
	hr.resq.ContentLength = int64(len(byts))
	hr.resq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return hr
}

// ToJSON .
func (hr *H2CRequest) ToJSON(obj interface{}) (r Response) {
	var body []byte
	r, body = hr.singleflightDo()
	if r.Error != nil {
		return
	}
	r.Error = json.Unmarshal(body, obj)
	if r.Error != nil {
		r.Error = fmt.Errorf("%s, body:%s", r.Error.Error(), string(body))
	}
	return
}

// ToString .
func (hr *H2CRequest) ToString() (value string, r Response) {
	var body []byte
	r, body = hr.singleflightDo()
	if r.Error != nil {
		return
	}
	value = string(body)
	return
}

// ToBytes .
func (hr *H2CRequest) ToBytes() (value []byte, r Response) {
	r, value = hr.singleflightDo()
	return
}

// ToXML .
func (hr *H2CRequest) ToXML(v interface{}) (r Response) {
	var body []byte
	r, body = hr.singleflightDo()
	if r.Error != nil {
		return
	}

	hr.httpRespone(&r)
	r.Error = xml.Unmarshal(body, v)
	return
}

// SetParam .
func (hr *H2CRequest) SetParam(key string, values ...interface{}) Request {
	for _, value := range values {
		hr.params.Add(key, fmt.Sprint(value))
	}
	return hr
}

// SetHeader .
func (hr *H2CRequest) SetHeader(header http.Header) Request {
	hr.resq.Header = header
	return hr
}

func (hr *H2CRequest) AddHeader(key, value string) Request {
	hr.resq.Header.Add(key, value)
	return hr
}

// URI .
func (hr *H2CRequest) URL() string {
	params := hr.params.Encode()
	if params != "" {
		return hr.url + "?" + params
	}
	return hr.url
}

func (hr *H2CRequest) singleflightDo() (r Response, body []byte) {
	if hr.singleflightKey == "" {
		r.Error = hr.do()
		if r.Error != nil {
			return
		}
		body = hr.body()
		hr.httpRespone(&r)
		return
	}

	data, _, _ := h2cclientGroup.Do(hr.singleflightKey, func() (interface{}, error) {
		var res Response
		res.Error = hr.do()
		if res.Error != nil {
			return &singleflightData{Res: res}, nil
		}

		hr.httpRespone(&res)
		return &singleflightData{Res: res, Body: hr.body()}, nil
	})

	sfdata := data.(*singleflightData)
	return sfdata.Res, sfdata.Body
}

func (hr *H2CRequest) do() (e error) {
	if hr.reqe != nil {
		return hr.reqe
	}

	u, e := url.Parse(hr.URL())
	if e != nil {
		return
	}
	if hr.ctx != nil {
		hr.resq = hr.resq.WithContext(hr.ctx)
	}

	hr.resq.URL = u
	var err error
	now := time.Now()
	defer func() {
		code := ""
		pro := ""
		if hr.resp != nil {
			code = strconv.Itoa(hr.resp.StatusCode)
			pro = hr.resp.Proto
		}
		if err != nil {
			code = "error"
		}
		PrometheusImpl.HttpClientWithLabelValues(u.Host, code, pro, hr.resq.Method, hr.name, now)
	}()
	hr.resp, err = h2cclient.Do(hr.resq)
	if err != nil {
		return err
	}
	code := hr.resp.StatusCode
	if code >= 400 && code <= 600 {
		return fmt.Errorf("The FastRequested URL returned error: %d", code)
	}
	return nil
}

func (hr *H2CRequest) body() (body []byte) {
	defer hr.resp.Body.Close()
	if hr.resp.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(hr.resp.Body)
		if err != nil {
			return
		}
		body, _ = ioutil.ReadAll(reader)
		return body
	}
	body, _ = ioutil.ReadAll(hr.resp.Body)
	return
}

func (hr *H2CRequest) httpRespone(httpRespone *Response) {
	httpRespone.StatusCode = hr.resp.StatusCode
	httpRespone.HTTP11 = false
	if hr.resp.ProtoMajor == 1 && hr.resp.ProtoMinor == 1 {
		httpRespone.HTTP11 = true
	}

	httpRespone.ContentLength = hr.resp.ContentLength
	httpRespone.ContentType = hr.resp.Header.Get("Content-Type")
	httpRespone.Header = hr.resp.Header
}

func (hr *H2CRequest) Singleflight(key ...interface{}) Request {
	hr.singleflightKey = fmt.Sprint(key...)
	return hr
}

func (fr *H2CRequest) SetName(name string) Request {
	fr.name = name
	return fr
}

func (fr *H2CRequest) GetName() string {
	return fr.name
}
