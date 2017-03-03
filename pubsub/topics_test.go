package pubsub

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

type MyClient struct {
	Uid     string
	Counter int
}

func (mc *MyClient) ID() string {
	return mc.Uid
}
func (mc *MyClient) OnMsg(message interface{}) {
	//fmt.Println(mc.Uid+"收到的消息:", message)
	mc.Counter++
}

type SlowClient struct {
	Uid     string
	Counter int
}

func (this *SlowClient) ID() string {
	return this.Uid
}
func (this *SlowClient) OnMsg(message interface{}) {
	fmt.Println("执行开始")
	time.Sleep(time.Second * 5)
	fmt.Println("执行结束")
	this.Counter++
}
func Test_Subscribe(t *testing.T) {
	clients := make([]*MyClient, 10)
	server := NewPubsub()
	for i := 0; i < len(clients); i++ {
		client := &MyClient{Uid: "client" + strconv.Itoa(i)}
		server.Subscribe(client.ID(), client.ID(), client.OnMsg)
		server.Subscribe("all", client.ID(), client.OnMsg)
		clients[i] = client
	}
	server.PushMessage("all", struct{}{})
	time.Sleep(time.Second)
	for i := 0; i < len(clients); i++ {
		server.Unsubscribe("all", clients[i].ID())
		server.Unsubscribe(clients[i].ID(), clients[i].ID())
	}
	for _, client := range clients {
		if client.Counter != 1 {
			t.Errorf("应该接受到1条消息，实际接受到%v条。", client.Counter)
		}
	}
}

// 注销测试
func Test_Unsubscribe(t *testing.T) {
	server := NewPubsub()
	if len(server.GetTopics()) != 0 {
		t.Error("初始化服务的主题数量不正确")
	}
	client := &SlowClient{Uid: "client" + strconv.Itoa(int(time.Now().Unix()))}
	//server.Subscribe(client, client.UID())
	server.Unsubscribe("all", client.ID())
	if len(server.GetTopics()) != 0 {
		t.Error("初始化服务的主题数量不正确")
	}
	server.Subscribe("all", client.ID(), client.OnMsg)
	server.PushMessage("all", "ok")
	if len(server.GetTopics()) != 1 {
		t.Error("初始化服务的主题数量不正确")
	}
}

// 注册\注销\发布各一次的性能
func BenchmarkServer_Subscribe(b *testing.B) {
	server := NewPubsub()
	for i := 0; i < b.N; i++ {
		client := &MyClient{Uid: "client" + strconv.Itoa(int(time.Now().Unix()))}
		//server.Subscribe(client, client.UID())
		server.Subscribe("all", client.ID(), client.OnMsg)
		server.PushMessage("all", "测试")
		server.Unsubscribe("all", client.ID())
		//server.Unsubscribe(client, client.UID())
	}
}

// 测试广播性能
func BenchmarkServer_PublishAll2(b *testing.B) {
	topicnamber := 1000
	clients := make([]*MyClient, topicnamber)
	server := NewPubsub()
	for i := 0; i < len(clients); i++ {
		clients[i] = &MyClient{Uid: "client" + strconv.Itoa(i)}
		server.Subscribe("all", clients[i].ID(), clients[i].OnMsg)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		j := i % topicnamber
		server.PushMessage("all", strconv.Itoa(j)) //,clients[i].UID+"发布消息")
	}
}

// 测试单点传送性能
func BenchmarkServer_PublishMessage(b *testing.B) {
	topicnamber := 10000
	clients := make([]*MyClient, topicnamber)
	server := NewPubsub()
	for i := 0; i < len(clients); i++ {
		clients[i] = &MyClient{Uid: "client" + strconv.Itoa(i)}
		//server.Subscribe(clients[j], "all")
		server.Subscribe(clients[i].ID(), clients[i].ID(), clients[i].OnMsg)
	}
	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		rem := j % topicnamber
		server.PushMessage("client", strconv.Itoa(rem))
	}
}
