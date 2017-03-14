// Copyright (C) 2012 Numerotron Inc.
// Use of this source code is governed by an MIT-style license
// that can be found in the LICENSE file.

// consistent包提供了一致性哈希的计算功能。
// 一致性哈希用于分布式请求的负载均衡。比如有三台机器，分别为cacheA,cacheB,cacheC，
// 我们可以通过提供的功能来分配大量用户的大批量请求。
//
// 我们可以通过New来创建一致性哈希，通过Add、Remove来增加、删除服务器，通过Get来获取提供服务的物理机。
// 需要注意的是,如果增删服务器，hash值将会重新计算（remap），会造成注册的服务器重建相关业务。
//
// 相关技术的材料，可以查看：
//	wikipedia:	http://en.wikipedia.org/wiki/Consistent_hashing
//	csdn:		http://blog.csdn.net/cywosp/article/details/23397179/
//	csdn:		http://blog.csdn.net/caigen1988/article/details/7708806
//	go语言实现:	https://segmentfault.com/a/1190000000414004
package consistent

import (
	"errors"
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

//虚拟服务器形成的环形数组，通过实现的Len等三个函数，来实现sort.Interface，从而方便实现相关排序功能。
type circle []uint32

// Len returns the length of the uints array.
func (x circle) Len() int { return len(x) }

// Less returns true if element i is less than element j.
func (x circle) Less(i, j int) bool { return x[i] < x[j] }

// Swap exchanges elements i and j.
func (x circle) Swap(i, j int) { x[i], x[j] = x[j], x[i] }

// ErrEmptyCircle is the error returned when trying to get an element when nothing has been added to hash.
var ErrEmptyCircle = errors.New("empty circle")

// 每个主机的虚拟节点数
// 设置为20的时候，测试用例不会报错。但修改之后，将会有部分报错，因为验证值是设定死了，会有部分用例报错。
var constReplicas = 20

// Consistent 通过成员变量，保留一致性哈希环的struct.
type Consistent struct {
	circle     circle            //环
	replicas   int               //每一个主机的复制份数,即一个主机对应的虚拟节点数
	virtualMap map[uint32]string //点到主机的映射
	members    map[string]bool   //主机列表
	sync.RWMutex
}

// NewConsistent 基于默认的replicas定义，创建一个新的对象。
//
// 要改变replicas的值，可以在添加服务器节点之前，通过SetReplicas方法修改。
func NewConsistent() *Consistent {
	c := &Consistent{
		circle:     circle{},
		replicas:   constReplicas,
		virtualMap: make(map[uint32]string),
		members:    make(map[string]bool),
	}
	return c
}

// 修改constent的replicas属性，务必在Add方法之前使用
func (c *Consistent) SetReplicas(replicasNum int) {
	c.replicas = replicasNum
}

// 获取一致性哈希中注册的主机节点数量
func (c *Consistent) GetMachineNum() int {
	return len(c.members)
}

// eltKey 为输入的值生成符合虚拟服务器命名规则的key值。
func (c *Consistent) eltKey(elt string, idx int) string {
	// return elt + "|" + strconv.Itoa(idx)
	return strconv.Itoa(idx) + elt
}

// Add 增加一个物理机到hash表中
func (c *Consistent) Add(elt string) {
	c.Lock()
	defer c.Unlock()
	c.add(elt)
}
func (c *Consistent) add(elt string) {
	//避免重复插入
	if _, ok := c.members[elt]; ok {
		return
	}
	for i := 0; i < c.replicas; i++ {
		c.virtualMap[c.hashKey(c.eltKey(elt, i))] = elt
	}
	c.members[elt] = true
	c.updateCircle()

}

// Remove removes an element from the hash.
func (c *Consistent) Remove(elt string) {
	c.Lock()
	defer c.Unlock()
	c.remove(elt)
}

// 调用前要加锁
func (c *Consistent) remove(elt string) {
	if _, ok := c.members[elt]; !ok {
		return
	}

	for i := 0; i < c.replicas; i++ {
		delete(c.virtualMap, c.hashKey(c.eltKey(elt, i)))
	}
	delete(c.members, elt)
	c.updateCircle()

}

// 批量设置物理服务器到hash中。如果输入值与hash已经存在的值不一致，则以输入值为准。
//	Set sets all the elements in the hash.  If there are existing elements not
//	present in elts, they will be removed.
func (c *Consistent) Set(elts []string) {
	c.Lock()
	defer c.Unlock()
	for k := range c.members {
		found := false
		for _, v := range elts {
			if k == v {
				found = true
				break
			}
		}
		if !found {
			c.remove(k)
		}
	}
	for _, v := range elts {
		_, exists := c.members[v]
		if exists {
			continue
		}
		c.add(v)
	}
}

// 以列表形式，获取所有的物理服务器
func (c *Consistent) Members() []string {
	c.RLock()
	defer c.RUnlock()
	var m []string
	for k := range c.members {
		m = append(m, k)
	}
	return m
}

// Get 按照顺时针取值原则，获取name的哈希值最接近的物理服务器，即circle[i-1]<hash(name)<=circle[i]，返回circle[i]对应的物理服务。
func (c *Consistent) Get(name string) (string, error) {
	c.RLock()
	defer c.RUnlock()
	if len(c.members) == 0 {
		return "", ErrEmptyCircle
	}
	key := c.hashKey(name)
	i := c.search(key)
	return c.virtualMap[c.circle[i]], nil
}

// 通过二分查找
func (c *Consistent) search(key uint32) (i int) {
	f := func(x int) bool {
		return c.circle[x] >= key //不能是‘<’符号，可以使用‘>’或者‘>=’
	}
	i = sort.Search(len(c.circle), f)
	if i >= len(c.circle) {
		i = 0
	}
	return
}

// GetTwo returns the two closest distinct elements to the name input in the circle.
func (c *Consistent) GetTwo(name string) (string, string, error) {
	c.RLock()
	defer c.RUnlock()
	if len(c.virtualMap) == 0 {
		return "", "", ErrEmptyCircle
	}
	key := c.hashKey(name)
	i := c.search(key)
	a := c.virtualMap[c.circle[i]]

	if c.GetMachineNum() == 1 {
		return a, "", nil
	}

	start := i
	var b string
	for i = start + 1; i != start; i++ {
		if i >= len(c.circle) {
			i = 0
		}
		b = c.virtualMap[c.circle[i]]
		if b != a {
			break
		}
	}
	return a, b, nil
}

// GetN returns the N closest distinct elements to the name input in the circle.
func (c *Consistent) GetN(name string, n int) ([]string, error) {
	c.RLock()
	defer c.RUnlock()

	if len(c.virtualMap) == 0 {
		return nil, ErrEmptyCircle
	}

	if c.GetMachineNum() < n {
		n = int(c.GetMachineNum())
	}

	var (
		key   = c.hashKey(name)
		i     = c.search(key)
		start = i
		res   = make([]string, 0, n)
		elem  = c.virtualMap[c.circle[i]]
	)

	res = append(res, elem)

	if len(res) == n {
		return res, nil
	}

	for i = start + 1; i != start; i++ {
		if i >= len(c.circle) {
			i = 0
		}
		elem = c.virtualMap[c.circle[i]]
		if !sliceContainsMember(res, elem) {
			res = append(res, elem)
		}
		if len(res) == n {
			break
		}
	}

	return res, nil
}

func (c *Consistent) hashKey(key string) uint32 {
	//	if len(key) < 64 {
	//		var scratch [64]byte
	//		copy(scratch[:], key)
	//		return crc32.ChecksumIEEE(scratch[:len(key)])
	//	}
	return crc32.ChecksumIEEE([]byte(key))
}

func (c *Consistent) updateCircle() {
	//	hashes := c.circle[:0]
	c.circle = circle{}
	//	//reallocate if we're holding on to too much (1/4th)
	//	if cap(c.circle)/(c.virtualNode*4) > len(c.virtualMap) {
	//		hashes = nil
	//	}
	for k := range c.virtualMap {
		c.circle = append(c.circle, k)
	}
	sort.Sort(c.circle)
}

func sliceContainsMember(set []string, member string) bool {
	for _, m := range set {
		if m == member {
			return true
		}
	}
	return false
}

// BUG(jc):没有实现热替换。某台服务器失效后，hash值需要全部重新计算，导致业务系统影响很大。
