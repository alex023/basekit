package basekit

import (
	"sync"
	"sync/atomic"
)

//Counter counter is a multi-thread safe counters
type Counter struct {
	mut     sync.Mutex
	currNum int64 //当前数量
	maxNum  int64 //最大数量
}

//AddOne 在原内部计数基础上，+1。
func (c *Counter) AddOne() int {
	c.mut.Lock()

	c.currNum++

	if c.maxNum < c.currNum {
		c.maxNum = c.currNum
	}
	c.mut.Unlock()

	return int(c.currNum)
}

//DecOne 在原内部计数基础上，-1。
func (c *Counter) DecOne() int {

	return int(atomic.AddInt64(&c.currNum, -1))

}

//Current 获取当前内部计数结果。
func (c *Counter) Current() int {
	return int(atomic.LoadInt64(&c.currNum))
}

//MaxNum 计数器生存周期内，最大的计数。
func (c *Counter) MaxNum() int {
	return int(atomic.LoadInt64(&c.maxNum))
}

//NewCounter counter constructor
func NewCounter() *Counter {
	return &Counter{}
}
