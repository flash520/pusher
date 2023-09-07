/**
 * @Author: koulei
 * @Description:
 * @File: handler
 * @Version: 1.0.0
 * @Date: 2023/9/4 20:13
 */

package pusher

import (
	"context"
	"strings"
	"sync"
)

type Handler interface {
	Name() string
	Handle(msg interface{}, user User)
	SetContext(ctx context.Context, cancelFunc context.CancelFunc)
	Clone() Handler
}

var defaultTopicHandler = &topicHandlers{
	container: map[string]Handler{},
}

type topicHandlers struct {
	mutex     sync.RWMutex
	container map[string]Handler
}

func (topic *topicHandlers) GetTopicHandler(name string) (Handler, bool) {
	topic.mutex.RLock()
	defer topic.mutex.RUnlock()
	handler, exists := topic.container[strings.ToLower(name)]
	if exists {
		return handler, true
	}
	return nil, false
}

func (topic *topicHandlers) Register(handler Handler) {
	topic.mutex.RLock()
	defer topic.mutex.RUnlock()

	topic.container[strings.ToLower(handler.Name())] = handler
}

func (topic *topicHandlers) UnRegister(name string) {
	topic.mutex.RLock()
	defer topic.mutex.RUnlock()

	delete(topic.container, strings.ToLower(name))
}
