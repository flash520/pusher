/**
 * @Author: koulei
 * @Description:
 * @File: response
 * @Version: 1.0.0
 * @Date: 2023/9/4 22:03
 */

package pusher

import (
	"encoding/json"
	"time"
)

type Response struct {
	Code      int         `json:"code"`
	Type      string      `json:"type"`
	RespName  string      `json:"name"`
	Body      interface{} `json:"body"`
	Error     string      `json:"error,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

func NewResponse(name string, data interface{}) *Response {
	var errMsg string
	var code int
	var ok bool
	var body interface{}
	err, ok := data.(error)
	if !ok {
		code = 1
		body = data
	} else {
		code = 0
		errMsg = err.Error()
	}
	return &Response{
		Code:      code,
		Type:      msgTypeMethod,
		RespName:  name,
		Body:      body,
		Error:     errMsg,
		Timestamp: time.Now().Unix(),
	}
}

func (resp *Response) SetName(name string) {
	resp.RespName = name
}

func (resp *Response) SetBody(data interface{}) {
	resp.Body = data
}

func (resp *Response) Name() string {
	return resp.RespName
}

func (resp *Response) First() bool {
	return true
}

func (resp *Response) Marshal() []byte {
	marshal, _ := json.Marshal(resp)
	return marshal
}
