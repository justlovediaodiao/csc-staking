package jsonrpc

import (
	"fmt"
	"net/http"
	"time"

	"csc/requests"
)

type request struct {
	JSONRPC string        `json:"jsonrpc"`
	METHOD  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int64         `json:"id"`
}

type response struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	Error   *Error      `json:"error"`
	ID      int64       `json:"id"`
}

// Error is jsonrpc response error
type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Error return error message
func (e Error) Error() string {
	return fmt.Sprintf("{code: %d, message: %s, data: %v}", e.Code, e.Message, e.Data)
}

// Response is jsonrpc response.
type Response struct {
	requests.Response
}

// Result get response result. v must be pointer or nil.
func (r *Response) Result(v interface{}) error {
	res := &response{Result: v}
	err := r.JSON(res)
	if err != nil {
		return err
	}
	if res.Error != nil {
		return res.Error
	}
	return nil
}

// Client is jsonrpc client with same timeout as requests.DefaultClient.
type Client struct {
	URL     string
	Headers http.Header
}

// Do do jsonrpc request.
func (c Client) Do(method string, args ...interface{}) (*Response, error) {
	body := &request{
		JSONRPC: "2.0",
		METHOD:  method,
		Params:  args,
		ID:      time.Now().Unix(),
	}
	req := requests.Post(c.URL)
	if c.Headers != nil {
		req.Headers = c.Headers
	}
	res, err := req.JSON(body).Result()
	if err != nil {
		return nil, err
	}
	return &Response{*res}, nil
}
