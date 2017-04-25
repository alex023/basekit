// Package pubsub is a simple subscription service module that provides asynchronous message distribution based on single computer memory
package pubsub

import (
	"github.com/alex023/basekit"
	"sync"
	"sync/atomic"
)

type message struct {
	topic string
	body  interface{}
}

// Pubsub is a  subscription service module
type Pubsub struct {
	sync.RWMutex
	dict     map[string]*Topic //map[topic.Name]*Channel
	wg       basekit.WaitWraper
	msgCache chan *message
	msgCount uint64
	exitFlag int32
}

// NewPubsub create a pubsub
func NewPubsub() *Pubsub {
	s := &Pubsub{
		dict:     make(map[string]*Topic),
		msgCache: make(chan *message, 1000),
	}
	s.wg.Wrap(func() { s.popMsg() })
	return s
}

//Subscribe 订阅主题，要确保输入的clientId唯一，避免不同客户端注册的时候采用同样的ClientId，否则会被替换。
func (s *Pubsub) Subscribe(topicName string, clientID string, callFunc func(msg interface{})) {
	s.RLock()
	ch, found := s.dict[topicName]
	s.RUnlock()
	//fmt.Println("那些订阅了的:", client.ID(), topicName)
	if !found {
		ch := NewTopic(topicName)
		ch.AddClient(clientID, callFunc)
		s.Lock()
		s.dict[topicName] = ch
		s.Unlock()
	} else {
		ch.AddClient(clientID, callFunc)
	}
}

//Unsubscribe 取消订阅。由于内部使用了waitgroup，在使用时，要特别小心：
//	1.订阅某个主题的handle，其内部不得直接调用Unsubscribe来注销同一主题。否则，如果该主题正好只有最后一个client，就会被阻塞。
//	2.如果确实需要，请加入：关键字 go。
func (s *Pubsub) Unsubscribe(topicName string, clientID string) {
	s.RLock()
	ch, found := s.dict[topicName]
	s.RUnlock()

	if found {
		if ch.DeleteClient(clientID) == 0 {
			ch.Close()
			s.Lock()
			delete(s.dict, topicName)
			s.Unlock()
		}
	}
}

// PushMessage asynchronous push a message
func (s *Pubsub) PushMessage(topicName string, m interface{}) {
	if atomic.LoadInt32(&s.exitFlag) == 1 {
		return
	}

	s.msgCache <- &message{topicName, m}

	atomic.AddUint64(&s.msgCount, 1)

}

func (s *Pubsub) popMsg() {
	for msg := range s.msgCache {
		s.wg.Wrap(func() { s.notifyMsg(msg.topic, msg.body) })
	}
	//s.wg.Wait()

}

//发布消息
func (s *Pubsub) notifyMsg(topicName string, message interface{}) bool {
	s.RLock()
	ch, found := s.dict[topicName]
	s.RUnlock()

	if !found {
		return false
	}
	return ch.NotifyMsg(message)
}

// Exiting returns a boolean indicating if topic is closed/exiting
func (s *Pubsub) Exiting() bool {
	return atomic.LoadInt32(&s.exitFlag) == 1
}

// Close safe exit service
func (s *Pubsub) Close() {
	if atomic.CompareAndSwapInt32(&s.exitFlag, 0, 1) {
		close(s.msgCache)
		s.wg.Wait()
	}
}

//GetTopics performs a thread safe operation to get all topics in subscription service module
func (s *Pubsub) GetTopics() []string {
	s.RLock()
	result := make([]string, len(s.dict))
	i := 0
	for topic := range s.dict {
		result[i] = topic
		i++
	}
	s.RUnlock()
	return result
}
