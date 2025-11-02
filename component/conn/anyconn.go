/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/1/1
   Description :
-------------------------------------------------
*/

package conn

import (
	"sync"

	"github.com/zly-app/zapp/consts"
	"github.com/zly-app/zapp/pkg/utils"
)

// 连接器
type AnyConn[T any] struct {
	wgs     map[string]*anyConnWaitGroup[T]
	mx      sync.RWMutex
	closeFn func(name string, conn T)
}

type anyConnWaitGroup[T any] struct {
	instance T
	e        error
	wg       sync.WaitGroup
}

func NewAnyConn[T any](closeFn func(name string, conn T)) *AnyConn[T] {
	return &AnyConn[T]{
		wgs:     make(map[string]*anyConnWaitGroup[T]),
		closeFn: closeFn,
	}
}

// 获取conn
func (c *AnyConn[T]) GetConn(creator func(name string) (T, error), name ...string) (T, error) {
	if len(name) == 0 {
		return c.getConn(creator, consts.DefaultComponentName)
	}
	return c.getConn(creator, name[0])
}

func (c *AnyConn[T]) getConn(creator func(name string) (T, error), name string) (T, error) {
	c.mx.RLock()
	wg, ok := c.wgs[name]
	c.mx.RUnlock()

	if ok {
		wg.wg.Wait()
		return wg.instance, wg.e
	}

	c.mx.Lock()

	// 再获取一次, 它可能在获取锁的过程中完成了
	if wg, ok = c.wgs[name]; ok {
		c.mx.Unlock()
		wg.wg.Wait()
		return wg.instance, wg.e
	}

	// 占位置
	wg = new(anyConnWaitGroup[T])
	wg.wg.Add(1)
	c.wgs[name] = wg
	c.mx.Unlock()

	var err error
	err = utils.Recover.WrapCall(func() error {
		wg.instance, err = creator(name)
		return err
	})

	// 如果出现错误, 删除占位
	if err != nil {
		wg.e = err
		wg.wg.Done()
		c.mx.Lock()
		delete(c.wgs, name)
		c.mx.Unlock()
		return *(new(T)), err
	}

	wg.wg.Done()
	return wg.instance, nil
}

// 关闭并移除conn
func (c *AnyConn[T]) Close(name string) {
	c.mx.Lock()
	wg, ok := c.wgs[name]
	if ok {
		delete(c.wgs, name)
	}
	c.mx.Unlock()

	if !ok {
		return
	}

	wg.wg.Wait()
	if c.closeFn != nil && wg.e == nil {
		c.closeFn(name, wg.instance)
	}
}

// 关闭并移除所有conn
func (c *AnyConn[T]) CloseAll() {
	c.mx.Lock()

	if c.closeFn == nil {
		c.wgs = make(map[string]*anyConnWaitGroup[T])
		c.mx.Unlock()
		return
	}

	fns := make([]func() error, 0, len(c.wgs))
	for k, v := range c.wgs {
		name := k
		wg := v
		fns = append(fns, func() error {
			wg.wg.Wait()
			if wg.e == nil {
				c.closeFn(name, wg.instance)
			}
			return nil
		})
	}
	c.wgs = make(map[string]*anyConnWaitGroup[T])
	c.mx.Unlock()

	_ = utils.Go.GoAndWait(fns...)
}
