package util

import "sync"

type CountableWaitGroup struct {
	wg    *sync.WaitGroup
	lock  *sync.Mutex
	count int
}

func NewCountableWaitGroup() *CountableWaitGroup {
	return &CountableWaitGroup{
		wg:    new(sync.WaitGroup),
		lock:  new(sync.Mutex),
		count: 0,
	}
}

func (c *CountableWaitGroup) Add(count int) {
	c.lock.Lock()
	c.count = c.count + count
	c.lock.Unlock()
	c.wg.Add(count)
}

func (c *CountableWaitGroup) Done() {
	c.lock.Lock()
	if c.count > 0 {
		c.count = c.count - 1
	}
	c.lock.Unlock()
	c.wg.Done()
}

func (c *CountableWaitGroup) Wait() {
	c.wg.Wait()
}

func (c *CountableWaitGroup) Count() int {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.count
}
