/**
 * @Author: koulei
 * @Description:
 * @File: handler
 * @Version: 1.0.0
 * @Date: 2023/9/6 14:14
 */

package main

import (
	"context"

	"github.com/flash520/pusher/pkg/pusher"
)

type Car struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	Number     int `json:"number"`
}

func (c *Car) Name() string {
	return "Car"
}

func (c *Car) Handle(msg interface{}, user pusher.User) {
	if user.First() {
		message := pusher.NewMessage(c.Name(), "这是首次加载数据", user.First())
		user.Write(message)
		return
	}
	c.Number++
	// err := fmt.Errorf("%s", "错啦~")
	// message := pusher.NewMessage(c.Name(), err, user.First())
	// user.Write(message)

	message := pusher.NewMessage(c.Name(), msg, user.First())
	user.Write(message)
}

func (c *Car) SetContext(ctx context.Context, cancelFunc context.CancelFunc) {
	c.ctx = ctx
	c.cancelFunc = cancelFunc
}

func (c *Car) Clone() pusher.Handler {
	return &Car{}
}

func init() {
	hub.TopicRegister(&Car{})
}