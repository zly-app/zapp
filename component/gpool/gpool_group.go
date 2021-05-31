/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/3/19
   Description :
-------------------------------------------------
*/

package gpool

import (
	"sync"
	"sync/atomic"

	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/pkg/utils"
)

// 协程池
type gpool struct {
	queue   chan *job // 任务队列
	wg      sync.WaitGroup
	isClose uint32
}

func newGPool(conf *GPoolConfig) core.IGPoolGroup {
	conf.check()
	g := &gpool{
		queue: make(chan *job, conf.JobQueueSize),
	}

	// 开始处理
	for i := 0; i < conf.ThreadCount; i++ {
		go g.start()
	}

	return g
}

func (g *gpool) start() {
	for job := range g.queue {
		g.process(job)
	}
}

func (g *gpool) process(job *job) {
	err := utils.Recover.WrapCall(job.fn)
	job.done <- err
	g.wg.Done()
}

// 异步执行
func (g *gpool) Go(fn func() error) <-chan error {
	g.wg.Add(1)
	job := newJob(fn)
	g.queue <- job
	return job.done
}

// 同步执行
func (g *gpool) GoSync(fn func() error) error {
	g.wg.Add(1)
	job := newJob(fn)
	g.queue <- job
	return <-job.done
}

// 等待所有任务执行完毕
func (g *gpool) Wait() {
	g.wg.Wait()
}

// 关闭, 关闭后禁止调用 Go 方法, 否则可能会产生panic
func (g *gpool) Close() {
	if atomic.CompareAndSwapUint32(&g.isClose, 0, 1) {
		close(g.queue)
	}
	g.wg.Wait()
}
