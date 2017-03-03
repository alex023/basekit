// package pubsub is a simple subscription service module that provides asynchronous message distribution based on single computer memory

package pubsub

import (
	"sync"
	"sync/atomic"
	"github.com/alex023/basetool"
)

type message struct {
	topic string
	body  interface{}
}

// Pubsub is a  subscription service module
type Pubsub struct {
	sync.RWMutex
	dict          map[string]*Topic //map[topic.Name]*Channel
	wg            basetool.WaitWraper
	exitChan      chan int
	messageCount  uint64
	exitFlag      int32
	memoryMsgChan chan *message
}
// NewPubsub create a pubsub
func NewPubsub() *Pubsub {
	s := &Pubsub{
		dict:          make(map[string]*Topic),
		memoryMsgChan: make(chan *message, 1000),
		exitChan:      make(chan int, 0),
	}
	s.wg.Wrap(func() { s.popMsg() })
	return s
}

//Subscribe 订阅主题，要确保输入的clientId唯一，避免不同客户端注册的时候采用同样的ClientId，否则会被替换。
func (s *Pubsub) Subscribe(topicName string, clientId string, callFunc func(msg interface{})) {
	s.RLock()
	ch, found := s.dict[topicName]
	s.RUnlock()
	//fmt.Println("那些订阅了的:", client.ID(), topicName)
	if !found {
		ch := NewTopic(topicName)
		ch.AddClient(clientId, callFunc)
		s.Lock()
		s.dict[topicName] = ch
		s.Unlock()
	} else {
		ch.AddClient(clientId, callFunc)
	}
}

//Unsubscribe 取消订阅。由于内部使用了waitgroup，在使用时，要特别小心：
//	1.订阅某个主题的handle，其内部不得直接调用Unsubscribe来注销同一主题。否则，如果该主题正好只有最后一个client，就会被阻塞。
//	2.如果确实需要，请加入：关键字 go。
func (s *Pubsub) Unsubscribe(topicName string, clientId string) {
	s.RLock()
	ch, found := s.dict[topicName]
	s.RUnlock()

	if found {
		if ch.DeleteClient(clientId) == 0 {
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

	s.memoryMsgChan <- &message{topicName, m}

	atomic.AddUint64(&s.messageCount, 1)

}

func (s *Pubsub) popMsg() {
	for {
		select {
		case msg := <-s.memoryMsgChan:
			s.notifyMsg(msg.topic, msg.body)
		case <-s.exitChan:
			goto Exit
		}
	}
Exit:
}

//发布消息
func (s *Pubsub) notifyMsg(topicName string, message interface{}) bool {
	s.RLock()
	ch, found := s.dict[topicName]
	s.RUnlock()

	if !found{
		return false
	}
	return ch.NotifyMsg(message)
}

// Exiting returns a boolean indicating if this topic is closed/exiting
func (s *Pubsub) Exiting() bool {
	return atomic.LoadInt32(&s.exitFlag) == 1
}

// Close safe exit service
func (s *Pubsub) Close() {
	if atomic.CompareAndSwapInt32(&s.exitFlag, 0, 1) {
		s.exitChan <- 1
		close(s.memoryMsgChan)
		s.wg.Wait()
	}
}
//GetTopics get all topics in subscription service module
func (s *Pubsub) GetTopics() []string {
	s.RLock()
	result := make([]string, len(s.dict))
	i := 0
	for topic, _ := range s.dict {
		result[i] = topic
		i++
	}
	s.RUnlock()
	return result
}
