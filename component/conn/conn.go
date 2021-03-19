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

	"go.uber.org/zap"

	"github.com/zly-app/zapp/consts"
	"github.com/zly-app/zapp/logger"
	"github.com/zly-app/zapp/pkg/utils"
)

type CreatorFunc = func(name string) (IInstance, error)

type IInstance interface {
	Close()
}

// 连接器
type Conn struct {
	wgs map[string]*connWaitGroup
	mx  sync.RWMutex
}

type connWaitGroup struct {
	instance IInstance
	e        error
	wg       sync.WaitGroup
}

func NewConn() *Conn {
	return &Conn{
		wgs: make(map[string]*connWaitGroup),
	}
}

// 获取实例
func (c *Conn) GetInstance(creator CreatorFunc, name ...string) IInstance {
	if len(name) == 0 {
		return c.getInstance(creator, consts.DefaultComponentName)
	}
	return c.getInstance(creator, name[0])
}

func (c *Conn) getInstance(creator CreatorFunc, name string) IInstance {
	c.mx.RLock()
	wg, ok := c.wgs[name]
	c.mx.RUnlock()

	if ok {
		wg.wg.Wait()
		if wg.e != nil {
			logger.Log.Panic(wg.e.Error(), zap.String("name", name))
		}
		return wg.instance
	}

	c.mx.Lock()

	// 再获取一次, 它可能在获取锁的过程中完成了
	if wg, ok = c.wgs[name]; ok {
		c.mx.Unlock()
		wg.wg.Wait()
		if wg.e != nil {
			logger.Log.Panic(wg.e.Error(), zap.String("name", name))
		}
		return wg.instance
	}

	// 占位置
	wg = new(connWaitGroup)
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
		logger.Log.Panic(wg.e.Error(), zap.String("name", name))
	}

	wg.wg.Done()
	return wg.instance
}

// 移除实例, 移除时会关闭示例
func (c *Conn) Remove(name string) {
	c.mx.Lock()
	defer c.mx.Unlock()
	wg, ok := c.wgs[name]
	if ok {
		wg.instance.Close()
		delete(c.wgs, name)
	}
}

// 关闭所有实例的链接
func (c *Conn) CloseAll() {
	c.mx.Lock()
	defer c.mx.Unlock()

	for _, wg := range c.wgs {
		if wg.instance != nil {
			wg.instance.Close()
		}
	}
	c.wgs = make(map[string]*connWaitGroup)
}
