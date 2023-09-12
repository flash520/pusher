/**
 * @Author: koulei
 * @Description:
 * @File: hub
 * @Version: 1.0.0
 * @Date: 2023/9/4 19:59
 */

package pusher

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type Reader interface {
	Name() string
	SetChannel(msg chan<- interface{})
	Start()
	Stop()
}

type Hub struct {
	mutex         sync.RWMutex
	clients       map[Client]struct{}
	handleRequest HandleRequest
	readerMutex   sync.RWMutex
	connectors    map[string]Reader
	event         chan interface{}
}

func NewHub() *Hub {
	hub := &Hub{
		clients:    make(map[Client]struct{}),
		connectors: make(map[string]Reader),
		event:      make(chan interface{}),
	}
	go hub.startReader()
	go hub.Run()
	return hub
}

func (h *Hub) ClientRegister(client Client) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if _, exists := h.clients[client]; !exists {
		h.clients[client] = struct{}{}
	}
}

func (h *Hub) ClientUnRegister(client Client) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if _, exists := h.clients[client]; exists {
		client.Close()
		delete(h.clients, client)
	}
}

func (h *Hub) TopicRegister(handler Handler) {
	defaultTopicHandler.Register(handler)
}

func (h *Hub) TopicUnRegister(handler Handler) {
	defaultTopicHandler.UnRegister(handler.Name())
}

// GetTopicHandler Fetch topic handler in repository
func (h *Hub) GetTopicHandler(name string) (Handler, bool) {
	return defaultTopicHandler.GetTopicHandler(name)
}

// HandleRequest Handle client request
func (h *Hub) HandleRequest(msg []byte, client Client) {
	if h.handleRequest == nil {
		h.defaultHandleRequest(msg, client)
		return
	}
	h.handleRequest(msg, client)
}

func (h *Hub) SetHandleRequest(handler HandleRequest) {
	h.handleRequest = handler
}

type ClientRequest struct {
	Method string   `json:"method"`
	Topics []string `json:"topics" binding:"required"`
}

func (h *Hub) defaultHandleRequest(msg []byte, client Client) {
	var request ClientRequest
	if err := json.Unmarshal(msg, &request); err != nil {
		resp := NewResponse("register", err)
		client.SendMessage(resp)
		return
	}

	if len(request.Topics) == 0 {
		err := fmt.Errorf("topic is empty")
		resp := NewResponse("register", err)
		client.SendMessage(resp)
		return
	}

	switch strings.ToLower(request.Method) {
	case "subscribe":
		for _, topic := range request.Topics {
			handler, b := h.GetTopicHandler(topic)
			if !b {
				err := fmt.Errorf("topic not found: %s", topic)
				resp := NewResponse("subscribe", err)
				client.SendMessage(resp)
				continue
			}

			newHandler := handler.Clone()
			if newHandler.Name() != handler.Name() {
				err := fmt.Errorf("handler clone failed: %s", topic)
				resp := NewResponse("subscribe", err)
				client.SendMessage(resp)
				continue
			}

			client.AppendTopicHandler(newHandler)
		}
	case "unsubscribe":
		client.DeleteTopicHandlers(request.Topics)
	default:
		err := fmt.Errorf("illegal method: %s", request.Method)
		resp := NewResponse(request.Method, err)
		client.SendMessage(resp)
	}
}

func (h *Hub) ReceiveChan() chan<- interface{} {
	return h.event
}

func (h *Hub) WriteEvent(event interface{}) {
	h.event <- event
}

func (h *Hub) Broadcast(msg interface{}) {
	for c := range h.clients {
		go c.HandleMessage(msg)
	}
}

func (h *Hub) SetReader(reader Reader) {
	h.readerMutex.RLock()
	defer h.readerMutex.RUnlock()

	_, exists := h.connectors[reader.Name()]
	if !exists {
		h.connectors[reader.Name()] = reader
		go reader.Start()
	}
}

func (h *Hub) startReader() {
	h.readerMutex.RLock()
	defer h.readerMutex.RUnlock()
	for _, reader := range h.connectors {
		go reader.Start()
	}
}

func (h *Hub) GetReader(name string) (Reader, bool) {
	h.readerMutex.RLock()
	defer h.readerMutex.RUnlock()

	if reader, exists := h.connectors[name]; exists {
		return reader, true
	}
	return nil, false
}

func (h *Hub) Run() {
	logrus.Infof("Starting Pusher Hub.")
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
	for {
		select {
		case msg := <-h.event:
			h.Broadcast(msg)
		case <-ticker.C:
			if len(h.clients) == 0 {
				continue
			}
			logrus.Infof("Current number of client connections: %d", len(h.clients))
		}
	}
}
