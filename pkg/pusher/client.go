/**
 * @Author: koulei
 * @Description:
 * @File: client
 * @Version: 1.0.0
 * @Date: 2023/9/4 20:00
 */

package pusher

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type HandleRequest func([]byte, Client)

type Client interface {
	SetHub(hub *Hub)
	User() User
	SetContext(ctx context.Context, cancelFunc context.CancelFunc)
	SendMessage(message Message)
	HandleMessage(msg interface{})
	AppendTopicHandler(handler Handler)
	DeleteTopicHandlers([]string)
	RemoteAddr() string
	Close()
	Run()
}

func NewClient(hub *Hub, conn *websocket.Conn) *client {
	ctx, cancelFunc := context.WithCancel(context.Background())
	c := &client{
		conn:    conn,
		hub:     hub,
		topics:  make(map[string]Handler),
		queue:   make(map[string]Message),
		msgChan: make(chan Message),
	}
	c.user = &userInfo{
		user:  nil,
		first: false,
		msg:   c.ReceiveChan(),
	}
	c.SetContext(ctx, cancelFunc)
	c.hub.ClientRegister(c)
	return c
}

type client struct {
	topicMutex sync.RWMutex
	hub        *Hub
	ctx        context.Context
	cancelFunc context.CancelFunc
	conn       *websocket.Conn
	topics     map[string]Handler
	queueMutex sync.RWMutex
	queue      map[string]Message
	msgChan    chan Message
	user       User
}

func (c *client) User() User {
	return c.user
}

func (c *client) SetContext(ctx context.Context, cancelFunc context.CancelFunc) {
	c.ctx = ctx
	c.cancelFunc = cancelFunc
}

func (c *client) SetHub(hub *Hub) {
	c.hub = hub
}

func (c *client) Run() {
	go c.readPump()
	go c.writePump()
	logrus.Infof("%s Connected", c.conn.RemoteAddr().String())
}

func (c *client) readPump() {
	defer func() { c.hub.ClientUnRegister(c) }()
	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPingHandler(c.Ping)
	c.conn.SetPongHandler(c.Pong)

	for {
		mt, msg, err := c.conn.ReadMessage()
		if err != nil {
			return
		}

		switch mt {
		case websocket.TextMessage:
			c.hub.HandleRequest(msg, c)
		case websocket.CloseMessage:
			return
		}
	}
}

func (c *client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	queueTicker := time.NewTicker(time.Second)
	defer func() {
		ticker.Stop()
		queueTicker.Stop()
	}()
	for {
		select {
		case <-ticker.C:
			c.heartbeat()
		case <-queueTicker.C:
			c.dispatch()
		case msg := <-c.msgChan:
			c.queue[msg.Name()] = msg
			if msg.First() {
				c.SendMessage(msg)
				c.queue[msg.Name()] = nil
			}
		priority:
			select {
			case <-queueTicker.C:
				c.dispatch()
			default:
				break priority
			}
		case <-c.ctx.Done():
			return
		}
	}
}

func (c *client) dispatch() {
	c.queueMutex.RLock()
	for key, msg := range c.queue {
		if c.queue[key] != nil {
			c.SendMessage(msg)
			c.queue[key] = nil
		}
	}
	c.queueMutex.RUnlock()
}

func (c *client) AppendTopicHandler(handler Handler) {
	c.queueMutex.RLock()

	c.topics[strings.ToLower(handler.Name())] = handler
	c.queueMutex.RUnlock()

	resp := NewResponse("register", fmt.Sprintf("topic %s subscribe success", handler.Name()))
	c.SendMessage(resp)
	c.user.SetFirst(true)
	handler.Handle(nil, c.user)
	c.user.SetFirst(false)
}

func (c *client) DeleteTopicHandlers(topics []string) {
	c.topicMutex.RLock()
	defer c.topicMutex.RUnlock()

	for _, topic := range topics {
		_, exists := c.topics[strings.ToLower(topic)]
		if exists {
			delete(c.topics, strings.ToLower(topic))
			delete(c.queue, strings.ToLower(topic))
			resp := NewResponse("unsubscribe", fmt.Sprintf("topic %s unsubscribe success", topic))
			c.SendMessage(resp)
		} else {
			err := fmt.Errorf("%s topic not found", topic)
			resp := NewResponse("unsubscribe", err)
			c.SendMessage(resp)
		}
	}
}

func (c *client) HandleMessage(msg interface{}) {
	for _, handler := range c.topics {
		go handler.Handle(msg, c.user)
	}
}

func (c *client) SendMessage(message Message) {
	if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		logrus.Errorf("Send Message Timeout, Error: %s", err.Error())
		return
	}
	if err := c.conn.WriteMessage(websocket.TextMessage, message.Marshal()); err != nil {
		logrus.Errorf("%s Send Message Error: %s", message.Name(), err.Error())
		return
	}
}

func (c *client) ReceiveChan() chan<- Message {
	return c.msgChan
}

func (c *client) heartbeat() {
	if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		logrus.Errorf("Send Message Timeout, Error: %s", err.Error())
		return
	}
	if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
		logrus.Errorf("Send Ping Error: %s", err.Error())
		return
	}
}

func (c *client) Ping(string) error {
	logrus.Infof("Receive Client Heartbeat Ping: %s", c.conn.RemoteAddr().String())
	if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		logrus.Errorf("Send Message Timeout, Error: %s", err.Error())
		return err
	}
	if err := c.conn.WriteMessage(websocket.PongMessage, nil); err != nil {
		logrus.Errorf("Send Ping Error: %s", err.Error())
		return err
	}

	return nil
}

func (c *client) Pong(string) error {
	logrus.Infof("Receive Client Heartbeat Pong: %s", c.conn.RemoteAddr().String())
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	return nil
}

func (c *client) RemoteAddr() string {
	return c.conn.RemoteAddr().String()
}

func (c *client) Close() {
	c.cancelFunc()
	_ = c.conn.Close()
	close(c.msgChan)
	logrus.Infof("%s Disconnected", c.conn.RemoteAddr().String())
}
