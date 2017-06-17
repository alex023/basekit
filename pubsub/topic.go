package pubsub

import (
	"github.com/alex023/basekit"
	"sync"
	"sync/atomic"
)

//Topic struct definition
type Topic struct {
	rwmut     sync.RWMutex
	Name      string
	wg        basekit.WaitWraper
	consumers map[string]func(interface{})
	msgCount  uint64
	exitFlag  int32
}

// NewTopic topic constructor
func NewTopic(topicName string) *Topic {
	return &Topic{Name: topicName, consumers: make(map[string]func(interface{}))}
}

//AddConsumer assign a new callback function to this topic
func (t *Topic) AddConsumer(clientID string, callFunc func(msg interface{})) bool {
	t.rwmut.Lock()
	t.consumers[clientID] = callFunc
	t.rwmut.Unlock()

	return true
}

//RmConsumer remove callback function by assigned clientid
func (t *Topic) RmConsumer(clientID string) int {
	t.rwmut.RLock()
	ret := len(t.consumers)
	t.rwmut.RUnlock()

	if ret > 0 {
		t.rwmut.Lock()
		delete(t.consumers, clientID)
		t.rwmut.Unlock()
		ret--
	}
	return ret
}

//NotifyMsg 向订阅了Topic的client发送消息
func (t *Topic) NotifyMsg(message interface{}) bool {
	t.rwmut.RLock()
	defer t.rwmut.RUnlock()
	if atomic.LoadInt32(&t.exitFlag) == 1 {
		return false
	}
	for _, client := range t.consumers {
		f := client
		t.wg.Wrap(func() { f(message) })
	}
	return true
}

// Close close mc topic until all messages have been sent to the registered client.
func (t *Topic) Close() {
	if !atomic.CompareAndSwapInt32(&t.exitFlag, 0, 1) {
		//add wg.Wait for every event should be sent to client when pubsub closing
		t.wg.Wait()
		t.consumers = nil
	}
}
