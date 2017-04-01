package basekit

import (
	"testing"
	"time"
)

func TestCounter_GetCurrentNum(t *testing.T) {
	var (
		countNum = 1000000
		c        = NewCounter()
	)
	for i := 0; i < countNum; i++ {
		c.AddOne()
	}
	if c.Current() != countNum {
		t.Errorf("result is bad:%d", c.MaxNum())
	}
}

func TestCounter_Current(t *testing.T) {
	var (
		countNum = 10000000
		c        = NewCounter()
	)
	go func() {
		for i := 0; i < countNum; i++ {
			c.AddOne()
		}
	}()
	go func() {
		for i := 0; i < countNum/2; i++ {
			c.AddOne()
		}
	}()
	time.Sleep(time.Second)
	x := c.Current()
	if x < 0 {
		t.Error("得到结果不正常，应该大于0")
	}
}

func BenchmarkCounter_AddOne(b *testing.B) {
	var (
		c = NewCounter()
	)
	for i := 0; i < b.N; i++ {
		c.AddOne()
	}
}
func BenchmarkCounter_DecOne(b *testing.B) {
	var (
		c = NewCounter()
	)
	for i := 0; i < b.N; i++ {
		c.DecOne()
	}
}