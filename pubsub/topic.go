package pubsub

import (
	"fmt"
	"sync"
	"sync/atomic"
	"github.com/alex023/basetool"
)

type Topic struct {
	sync.RWMutex
	wg           basetool.WaitWraper
	Name         string
	clients      map[string]func(interface{})
	messagecount uint64
	exitFlag     int32
}

func NewTopic(topicName string) *Topic {
	return &Topic{Name: topicName, clients: make(map[string]func(interface{}))}
}
func (t *Topic) AddClient(clientId string, callFunc func(msg interface{})) bool {
	//t.RLock()
	//_, found := t.clients[client.ID()]
	//t.RUnlock()
	//
	//if !found {
	t.Lock()
	t.clients[clientId] = callFunc
	t.Unlock()
	//}
	return true

}
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
	if t.closing(){
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
// Close close this topic until all messages have been sent to the registered client.
func (t *Topic) Close() {
	if !atomic.CompareAndSwapInt32(&t.exitFlag, 0, 1) {
		return
	}
	//等待正在执行的广播消息完成，通过wait确保注册方法的执行
	t.wg.Wait()

}
