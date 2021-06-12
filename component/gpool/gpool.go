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

	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/pkg/utils"
)

// 协程池
type gpool struct {
	workerQueue chan *worker  // 工人队列
	jobQueue    chan *job     // 任务队列
	stop        chan struct{} // 停止信号, 同步通道

	wg sync.WaitGroup
}

func NewGPool(conf *GPoolConfig) core.IGPool {
	conf.check()
	g := &gpool{
		workerQueue: make(chan *worker, conf.ThreadCount),
		jobQueue:    make(chan *job, conf.JobQueueSize),
		stop:        make(chan struct{}),
	}

	for i := 0; i < conf.ThreadCount; i++ {
		worker := newWorker(g.workerQueue)
		worker.Ready()
		g.workerQueue <- worker
	}

	go g.dispatch()

	return g
}

// 为工人派遣任务
func (g *gpool) dispatch() {
	for {
		select {
		case job := <-g.jobQueue:
			worker := <-g.workerQueue
			worker.Do(job)
		case <-g.stop:
			for i := 0; i < cap(g.workerQueue); i++ {
				worker := <-g.workerQueue
				worker.Stop()
			}

			g.stop <- struct{}{}
			return
		}
	}
}

// 异步执行
func (g *gpool) Go(fn func() error) <-chan error {
	job := g.newJob(fn)
	g.wg.Add(1)
	g.jobQueue <- job
	return job.done
}

// 同步执行
func (g *gpool) GoSync(fn func() error) error {
	return <-g.Go(fn)
}

// 尝试异步执行, 如果任务队列已满则返回false
func (g *gpool) TryGo(fn func() error) (ch <-chan error, ok bool) {
	job := g.newJob(fn)
	select {
	case g.jobQueue <- job:
		g.wg.Add(1)
		return job.done, true
	default:
		return nil, false
	}
}

// 尝试同步执行, 如果任务队列已满则返回false
func (g *gpool) TryGoSync(fn func() error) (result error, ok bool) {
	ch, ok := g.TryGo(fn)
	if !ok {
		return nil, false
	}
	return <-ch, true
}

// 等待所有任务结束
func (g *gpool) Wait() {
	g.wg.Wait()
}

// 关闭
//
// 命令所有没有收到任务的工人立即停工, 收到任务的工人完成当前任务后停工, 不管任务队列是否清空
func (g *gpool) Close() {
	g.stop <- struct{}{}
	<-g.stop
}

func (g *gpool) newJob(fn func() error) *job {
	return newJob(func() error {
		defer g.wg.Done()
		return utils.Recover.WrapCall(fn)
	})
}
