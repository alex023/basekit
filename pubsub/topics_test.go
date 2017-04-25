package pubsub

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"
)

type MyClient struct {
	UID     string
	Counter int
}

func (mc *MyClient) ID() string {
	return mc.UID
}
func (mc *MyClient) OnMsg(message interface{}) {
	//fmt.Println(mc.UID+"收到的消息:", message)
	mc.Counter++
}

type SlowClient struct {
	mut     sync.Mutex
	UID     string
	Counter int
}

func (sc *SlowClient) ID() string {
	return sc.UID
}
func (sc *SlowClient) OnMsg(message interface{}) {
	sc.mut.Lock()
	defer sc.mut.Unlock()

	time.Sleep(time.Microsecond*100)
	sc.Counter++
}
func Test_Subscribe(t *testing.T) {
	clients := make([]*MyClient, 10)
	center := NewPubsub()
	for i := 0; i < len(clients); i++ {
		client := &MyClient{UID: "client" + strconv.Itoa(i)}
		center.Subscribe(client.ID(), client.ID(), client.OnMsg)
		center.Subscribe("all", client.ID(), client.OnMsg)
		clients[i] = client
	}
	center.PushMessage("all", struct{}{})
	time.Sleep(time.Second)
	for i := 0; i < len(clients); i++ {
		center.Unsubscribe("all", clients[i].ID())
		center.Unsubscribe(clients[i].ID(), clients[i].ID())
	}
	for _, client := range clients {
		if client.Counter != 1 {
			t.Errorf("应该接受到1条消息，实际接受到%v条。", client.Counter)
		}
	}
}

// 注销测试
func Test_Unsubscribe(t *testing.T) {
	center := NewPubsub()
	if len(center.GetTopics()) != 0 {
		t.Error("初始化服务的主题数量不正确")
	}
	client := &SlowClient{UID: "client" + strconv.Itoa(int(time.Now().Unix()))}
	//server.Subscribe(client, client.UID())
	center.Unsubscribe("all", client.ID())
	if len(center.GetTopics()) != 0 {
		t.Error("初始化服务的主题数量不正确")
	}
	center.Subscribe("all", client.ID(), client.OnMsg)
	center.PushMessage("all", "ok")
	if len(center.GetTopics()) != 1 {
		t.Error("初始化服务的主题数量不正确")
	}
}
func TestPubsub_Close(t *testing.T) {
	center := NewPubsub()
	client := &SlowClient{UID: "client" + strconv.Itoa(int(time.Now().Unix()))}

	center.Subscribe("all", client.ID(), client.OnMsg)
	for i := 0; i < 10; i++ {
		center.PushMessage("all", "ok")
	}
	center.Close()
	if client.Counter == 1000 {
		t.Error("客户端响应数量存在问题！")
	}
}

// 注册\注销\发布各一次的性能
func BenchmarkServer_Subscribe(b *testing.B) {
	center := NewPubsub()
	for i := 0; i < b.N; i++ {
		client := &MyClient{UID: "client" + strconv.Itoa(int(time.Now().Unix()))}
		//center.Subscribe(client, client.UID())
		center.Subscribe("all", client.ID(), client.OnMsg)
		center.PushMessage("all", "测试")
		center.Unsubscribe("all", client.ID())
		//center.Unsubscribe(client, client.UID())
	}
}

// 测试广播性能
func BenchmarkServer_PublishAll2(b *testing.B) {
	topicnamber := 1000
	clients := make([]*MyClient, topicnamber)
	center := NewPubsub()
	for i := 0; i < len(clients); i++ {
		clients[i] = &MyClient{UID: "client" + strconv.Itoa(i)}
		center.Subscribe("all", clients[i].ID(), clients[i].OnMsg)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		j := i % topicnamber
		center.PushMessage("all", strconv.Itoa(j)) //,clients[i].UID+"发布消息")
	}
}

// 测试单点传送性能
func BenchmarkServer_PublishMessage(b *testing.B) {
	topicnamber := 10000
	clients := make([]*MyClient, topicnamber)
	center := NewPubsub()
	for i := 0; i < len(clients); i++ {
		clients[i] = &MyClient{UID: "client" + strconv.Itoa(i)}
		//center.Subscribe(clients[j], "all")
		center.Subscribe(clients[i].ID(), clients[i].ID(), clients[i].OnMsg)
	}
	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		rem := j % topicnamber
		center.PushMessage("client", strconv.Itoa(rem))
	}
}
