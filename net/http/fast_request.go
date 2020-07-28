package http

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"golang.org/x/sync/singleflight"
)

var (
	Fastclient      *http.Client
	fastclientGroup singleflight.Group
)

func init() {
	NewFastClient(15*time.Second, 512, false)
}

// Newhttpclient .
func NewFastClient(rwTimeout time.Duration, MaxIdleConns int, disableKeepAlives bool) {
	tran := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: 15 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          MaxIdleConns,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   100,
		DisableKeepAlives:     disableKeepAlives,
	}

	Fastclient = &http.Client{
		Transport: tran,
		Timeout:   rwTimeout,
	}
}

// NewFastRequest .
func NewFastRequest(uri ...string) Request {
	result := new(FastRequest)
	req := &http.Request{
		Header: make(http.Header),
	}
	result.resq = req
	result.params = make(url.Values)
	if len(uri) > 0 {
		result.url = uri[0]
	}
	return result
}

// FastRequest .
type FastRequest struct {
	resq            *http.Request
	resp            *http.Response
	reqe            error
	params          url.Values
	url             string
	singleflightKey string
	name            string
	responseError   error
	stop            bool
}

// SetURL .
func (fr *FastRequest) SetURL(uri string) Request {
	fr.url = uri
	return fr
}

// SetContext .
func (fr *FastRequest) SetContext(ctx context.Context) Request {
	fr.resq = fr.resq.WithContext(ctx)
	return fr
}

// Post .
func (fr *FastRequest) Post() Request {
	fr.resq.Method = "POST"
	return fr
}

// Put .
func (fr *FastRequest) Put() Request {
	fr.resq.Method = "PUT"
	return fr
}

// Get .
func (fr *FastRequest) Get() Request {
	fr.resq.Method = "GET"
	return fr
}

// Get .
func (fr *FastRequest) Head() Request {
	fr.resq.Method = "HEAD"
	return fr
}

// // Delete .
func (fr *FastRequest) Delete() Request {
	fr.resq.Method = "DELETE"
	return fr
}

// SetJSONBody .
func (fr *FastRequest) SetJSONBody(obj interface{}) Request {
	byts, e := json.Marshal(obj)
	if e != nil {
		fr.reqe = e
		return fr
	}

	fr.resq.Body = ioutil.NopCloser(bytes.NewReader(byts))
	fr.resq.ContentLength = int64(len(byts))
	fr.resq.Header.Set("Content-Type", "application/json")
	return fr
}

// SetBody .
func (fr *FastRequest) SetBody(byts []byte) Request {
	fr.resq.Body = ioutil.NopCloser(bytes.NewReader(byts))
	fr.resq.ContentLength = int64(len(byts))
	fr.resq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return fr
}

// ToJSON .
func (fr *FastRequest) ToJSON(obj interface{}) (r Response) {
	var body []byte
	r, body = fr.singleflightDo()
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
func (fr *FastRequest) ToString() (value string, r Response) {
	var body []byte
	r, body = fr.singleflightDo()
	if r.Error != nil {
		return
	}
	value = string(body)
	return
}

// ToBytes .
func (fr *FastRequest) ToBytes() (value []byte, r Response) {
	r, value = fr.singleflightDo()
	return
}

// ToXML .
func (fr *FastRequest) ToXML(v interface{}) (r Response) {
	var body []byte
	r, body = fr.singleflightDo()
	if r.Error != nil {
		return
	}

	fr.httpRespone(&r)
	r.Error = xml.Unmarshal(body, v)
	return
}

// SetParam .
func (fr *FastRequest) SetParam(key string, values ...interface{}) Request {
	for _, value := range values {
		fr.params.Add(key, fmt.Sprint(value))
	}
	return fr
}

// SetHeader .
func (fr *FastRequest) SetHeader(header http.Header) Request {
	fr.resq.Header = header
	return fr
}

func (fr *FastRequest) AddHeader(key, value string) Request {
	fr.resq.Header.Add(key, value)
	return fr
}

// URI .
func (fr *FastRequest) URL() string {
	params := fr.params.Encode()
	if params != "" {
		return fr.url + "?" + params
	}
	return fr.url
}

func (fr *FastRequest) singleflightDo() (r Response, body []byte) {
	if fr.singleflightKey == "" {
		r.Error = fr.prepare()
		if r.Error != nil {
			return
		}
		handle(fr)
		r.Error = fr.responseError
		if r.Error != nil {
			return
		}
		body = fr.body()
		fr.httpRespone(&r)
		return
	}

	data, _, _ := fastclientGroup.Do(fr.singleflightKey, func() (interface{}, error) {
		var res Response
		res.Error = fr.prepare()
		if r.Error != nil {
			return &singleflightData{Res: res}, nil
		}
		handle(fr)
		res.Error = fr.responseError
		if res.Error != nil {
			return &singleflightData{Res: res}, nil
		}

		fr.httpRespone(&res)
		return &singleflightData{Res: res, Body: fr.body()}, nil
	})

	sfdata := data.(*singleflightData)
	return sfdata.Res, sfdata.Body
}

func (fr *FastRequest) do() (e error) {
	if fr.reqe != nil {
		return fr.reqe
	}

	u, e := url.Parse(fr.URL())
	if e != nil {
		return
	}
	fr.resq.URL = u
	var err error
	now := time.Now()
	defer func() {
		code := ""
		pro := ""
		if fr.resp != nil {
			code = strconv.Itoa(fr.resp.StatusCode)
			pro = fr.resp.Proto
		}
		if err != nil {
			code = "error"
		}
		PrometheusImpl.HttpClientWithLabelValues(u.Host, code, pro, fr.resq.Method, fr.name, now)
	}()
	fr.resp, err = Fastclient.Do(fr.resq)
	if err != nil {
		return err
	}
	code := fr.resp.StatusCode
	if code >= 400 && code <= 600 {
		return fmt.Errorf("The FastRequested URL returned error: %d", code)
	}
	return nil
}

func (fr *FastRequest) body() (body []byte) {
	defer fr.resp.Body.Close()
	if fr.resp.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(fr.resp.Body)
		if err != nil {
			return
		}
		body, _ = ioutil.ReadAll(reader)
		return body
	}
	body, _ = ioutil.ReadAll(fr.resp.Body)
	return
}

func (fr *FastRequest) httpRespone(httpRespone *Response) {
	httpRespone.StatusCode = fr.resp.StatusCode
	httpRespone.HTTP11 = false
	if fr.resp.ProtoMajor == 1 && fr.resp.ProtoMinor == 1 {
		httpRespone.HTTP11 = true
	}

	httpRespone.ContentLength = fr.resp.ContentLength
	httpRespone.ContentType = fr.resp.Header.Get("Content-Type")
	httpRespone.Header = fr.resp.Header
}

func (fr *FastRequest) Singleflight(key ...interface{}) Request {
	fr.singleflightKey = fmt.Sprint(key...)
	return fr
}

func (fr *FastRequest) SetName(name string) Request {
	fr.name = name
	return fr
}

func (fr *FastRequest) GetName() string {
	return fr.name
}

func (req *FastRequest) prepare() (e error) {
	if req.reqe != nil {
		return req.reqe
	}

	u, e := url.Parse(req.URL())
	if e != nil {
		return
	}
	req.resq.URL = u
	return
}

func (req *FastRequest) Next() {
	req.responseError = req.do()
}

func (req *FastRequest) Stop(e ...error) {
	req.stop = true
	if len(e) > 0 {
		req.responseError = e[0]
		return
	}
	req.responseError = errors.New("Middleware stop")
}

func (req *FastRequest) getStop() bool {
	return req.stop
}

func (req *FastRequest) GetRequest() *http.Request {
	return req.resq
}

func (req *FastRequest) GetRespone() (*http.Response, error) {
	return req.resp, req.responseError
}
