package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type request struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
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
	data []byte
}

// Result get response result. v must be pointer or nil.
func (r *Response) Result(v interface{}) error {
	res := response{Result: v}
	err := json.Unmarshal(r.data, &res)
	if err != nil {
		return err
	}
	if res.Error != nil {
		return res.Error
	}
	return nil
}

// Client is jsonrpc client with same timeout as requests.DefaultClient.
type RPCClient struct {
	URL    string
	Client *http.Client
}

// Do do jsonrpc request.
func (c RPCClient) Do(method string, args ...interface{}) (*Response, error) {
	req := &request{
		JSONRPC: "2.0",
		Method:  method,
		Params:  args,
		ID:      time.Now().Unix(),
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	res, err := c.Client.Post(c.URL, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("http status code: %d", res.StatusCode)
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return &Response{data}, nil
}
