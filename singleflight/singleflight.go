// copy from github.com/golang/groupcache/singleflight.go

// Package singleflight provide  a duplicate function call suppression mechisam.
package singleflight

import "sync"

//call is an in-flight or completed Do call
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

// Group represents a class of work and forms a namespace in which
// units of work can be executed with duplicate suppression.
type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

// Do executes and returns the results of the given function, making
// sure that only one execution is in-flight for a given key at a
// time. If a duplicate comes in, the duplicate caller waits for the
// original to complete and receives the same results.
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, founded := g.m[key]; founded {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, nil
	}
	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}

var defaultGroup = new(Group)
//Do  runs the defaultGroup on a single Group.
func Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	return defaultGroup.Do(key, fn)
}
