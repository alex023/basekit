package basekit

import (
	"sync"
)

//Counter counter is a multi-thread safe counters
type Counter struct {
	mut     sync.Mutex
	currNum int //当前数量
	maxNum  int //最大数量
}

//AddOne 在原内部计数基础上，+1。
func (c *Counter) AddOne() int {
	c.mut.Lock()
	defer c.mut.Unlock()

	c.currNum++

	if c.maxNum < c.currNum {
		c.maxNum = c.currNum
	}

	return c.currNum
}

//DecOne 在原内部计数基础上，-1。
func (c *Counter) DecOne() int {
	c.mut.Lock()
	defer c.mut.Unlock()

	c.currNum--

	return c.currNum
}

//Current 获取当前内部计数结果。
func (c *Counter) Current() int {
	return c.currNum
}

//MaxNum 计数器生存周期内，最大的计数。
func (c *Counter) MaxNum() int {
	return c.maxNum
}

//NewCounter counter constructor
func NewCounter() *Counter {
	return &Counter{}
}
