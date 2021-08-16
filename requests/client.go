package requests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// ErrStatusCode is http status code error.
type ErrStatusCode struct {
	StatusCode int
	Status     string
}

// Error return error message.
func (e ErrStatusCode) Error() string {
	return fmt.Sprintf("%d %s", e.StatusCode, e.Status)
}

// DefaultClient is http client with default timeout 15s.
var DefaultClient = &http.Client{
	Timeout: time.Second * 15,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
	},
}

// Value is a string map.
type Value map[string]string

// Encode encode to `URL encoded` string.
func (v Value) Encode() string {
	var buf strings.Builder
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := v[k]
		if buf.Len() > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(url.QueryEscape(k))
		buf.WriteByte('=')
		buf.WriteString(url.QueryEscape(v))

	}
	return buf.String()
}

// Request is http request config.
type Request struct {
	Method  string
	URL     string
	Args    Value
	Headers http.Header
	Body    io.Reader
	err     error
}

// Arg add args to request.
func (r *Request) Arg(v Value) *Request {
	if r.Args == nil {
		r.Args = v
	} else {
		for key, val := range v {
			r.Args[key] = val
		}
	}
	return r
}

// Header add headers to request.
func (r *Request) Header(v Value) *Request {
	if r.Headers == nil {
		r.Headers = http.Header{}
	}
	for k, v := range v {
		r.Headers.Add(k, v)
	}
	return r
}

// JSON add json body and header to request.
func (r *Request) JSON(v interface{}) *Request {
	body, err := json.Marshal(v)
	if err != nil {
		r.err = err
	} else {
		r.Body = bytes.NewReader(body)
		if r.Headers == nil {
			r.Headers = http.Header{}
		}
		r.Headers.Set("Content-Type", "application/json")
	}
	return r
}

// Form add form body and header to request.
func (r *Request) Form(v Value) *Request {
	r.Body = strings.NewReader(v.Encode())
	if r.Headers == nil {
		r.Headers = http.Header{}
	}
	r.Headers.Set("Content-Type", "application/x-www-form-urlencoded")
	return r
}

// Data add body to request.
func (r *Request) Data(body io.Reader) *Request {
	r.Body = body
	return r
}

// Prepare prepare and return http request.
func (r *Request) Prepare() (*http.Request, error) {
	if r.err != nil {
		return nil, r.err
	}
	var url = r.URL
	if r.Args != nil {
		url = fmt.Sprintf("%s?%s", r.URL, r.Args.Encode())
	}
	req, err := http.NewRequest(r.Method, url, r.Body)
	if err != nil {
		return nil, err
	}
	req.Header = r.Headers
	return req, nil
}

// Reslove send request by http.Client.
// If http.Client is nil, use DefaultClient.
func (r *Request) Reslove(c *http.Client) (*Response, error) {
	if c == nil {
		c = DefaultClient
	}
	req, err := r.Prepare()
	if err != nil {
		return nil, err
	}
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	content, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return &Response{Raw: *res, content: content}, nil
}

// Result send request by DefaultClient.
func (r *Request) Result() (*Response, error) {
	return r.Reslove(nil)
}

// Response is http response.
type Response struct {
	Raw     http.Response
	content []byte
}

func (r *Response) checkStatusCode() error {
	if r.Raw.StatusCode != http.StatusOK {
		return ErrStatusCode{r.Raw.StatusCode, r.Raw.Status}
	}
	return nil
}

// Content return response content.
func (r *Response) Content() ([]byte, error) {
	if err := r.checkStatusCode(); err != nil {
		return nil, err
	}
	return r.content, nil
}

// Text parse response content to string.
func (r *Response) Text() (string, error) {
	if err := r.checkStatusCode(); err != nil {
		return "", err
	}
	return string(r.content), nil
}

// JSON parse response content to json object. v must be a pointer.
func (r *Response) JSON(v interface{}) error {
	if err := r.checkStatusCode(); err != nil {
		return err
	}
	return json.Unmarshal(r.content, v)
}

// Get return a request with GET method.
func Get(url string) *Request {
	return &Request{
		Method: "GET",
		URL:    url,
	}
}

// Post return a request with POST method.
func Post(url string) *Request {
	return &Request{
		Method: "POST",
		URL:    url,
	}
}

// Put return a request with PUT method.
func Put(url string) *Request {
	return &Request{
		Method: "PUT",
		URL:    url,
	}
}

// Patch return a request with PATCH method.
func Patch(url string) *Request {
	return &Request{
		Method: "PATCH",
		URL:    url,
	}
}

// Delete return a request with DELETE method.
func Delete(url string) *Request {
	return &Request{
		Method: "DELETE",
		URL:    url,
	}
}
