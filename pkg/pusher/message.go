/**
 * @Author: koulei
 * @Description:
 * @File: message
 * @Version: 1.0.0
 * @Date: 2023/9/4 21:48
 */

package pusher

import (
	"encoding/json"
	"time"
)

type Message interface {
	Name() string
	Marshal() []byte
	First() bool
}

type message struct {
	Code      int         `json:"code"`
	Type      string      `json:"type"`
	Topic     string      `json:"name"`
	Body      interface{} `json:"body"`
	Error     string      `json:"error,omitempty"`
	Timestamp int64       `json:"timestamp"`
	first     bool
}

func NewMessage(name string, data interface{}, first bool) Message {
	var code int
	var errMsg string
	var body interface{}
	err, ok := data.(error)
	if ok {
		code = 0
		errMsg = err.Error()
	} else {
		code = 1
		body = data
	}
	return &message{
		Code:      code,
		Type:      msgTypeData,
		Topic:     name,
		Body:      body,
		first:     first,
		Error:     errMsg,
		Timestamp: time.Now().Unix(),
	}
}

func (msg *message) Name() string {
	return msg.Topic
}

func (msg *message) First() bool {
	return msg.first
}

func (msg *message) Marshal() []byte {
	marshal, _ := json.Marshal(msg)
	return marshal
}
