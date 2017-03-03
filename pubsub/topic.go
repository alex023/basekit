package pubsub

import (
	"fmt"
	"github.com/alex023/basekit"
	"sync"
	"sync/atomic"
)

type Topic struct {
	sync.RWMutex
	wg           basekit.WaitWraper
	Name         string
	clients      map[string]func(interface{})
	messagecount uint64
	exitFlag     int32
}
// NewTopic create a topic by assigned name
func NewTopic(topicName string) *Topic {
	return &Topic{Name: topicName, clients: make(map[string]func(interface{}))}
}
//AddClient assign a new callback function
func (t *Topic) AddClient(clientId string, callFunc func(msg interface{})) bool {
	t.Lock()
	t.clients[clientId] = callFunc
	t.Unlock()

	return true
}
//DeleteClient remove callback function by assigned clientid
func (t *Topic) DeleteClient(clientId string) int {
	t.RLock()
	ret := len(t.clients)
	t.RUnlock()

	if ret > 0 {
		t.Lock()
		delete(t.clients, clientId)
		t.Unlock()
		ret--
	}
	return ret
}

func (t *Topic) NotifyMsg(message interface{}) bool {
	t.RLock()
	defer t.RUnlock()
	if t.closing() {
		return false
	}
	for _, client := range t.clients {
		f := client
		t.wg.Wrap(func() { f(message) })
	}
	return true
}

func (t *Topic) ReplyMsg(message interface{}) {
	t.wg.Wrap(func() { fmt.Println(message) })
}

func (t *Topic) closing() bool {
	return atomic.LoadInt32(&t.exitFlag) == 1
}

// Close close mc topic until all messages have been sent to the registered client.
func (t *Topic) Close() {
	if !atomic.CompareAndSwapInt32(&t.exitFlag, 0, 1) {
		return
	}
	//等待正在执行的广播消息完成，通过wait确保注册方法的执行
	t.wg.Wait()

}
