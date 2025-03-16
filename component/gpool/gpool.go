/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/3/19
   Description :
-------------------------------------------------
*/

package gpool

import (
	"errors"
	"sync"

	"github.com/zly-app/zapp/core"
)

var ErrGPoolClosed = errors.New("gpool closed")

// 协程池
type gpool struct {
	workerQueue chan *worker  // 工人队列
	jobQueue    chan *job     // 任务队列
	stop        chan struct{} // 停止信号, 同步通道

	wg        sync.WaitGroup
	onceClose sync.Once
}

func NewGPool(conf *GPoolConfig) core.IGPool {
	conf.check()
	if conf.ThreadCount < 0 {
		return NewNoPool()
	}

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
	var worker *worker
	var stop bool
	for !stop {
		if worker == nil {
			select {
			case w := <-g.workerQueue:
				worker = w
			case <-g.stop:
				stop = true
				continue
			}
		}

		select {
		case job := <-g.jobQueue:
			worker.Do(job)
			worker = nil
		case <-g.stop:
			stop = true
		}
	}

	// 释放worker
	if worker != nil {
		worker.Do(nil) // 让这个工人回到池
	}
	for i := 0; i < cap(g.workerQueue); i++ {
		w := <-g.workerQueue
		w.Stop()
	}
	g.workerQueue = nil

	// 释放job
	jobLen := len(g.jobQueue)
	g.jobQueue = nil
	for i := 0; i < jobLen; i++ {
		g.wg.Done()
	}

	g.stop <- struct{}{}
}

// 异步执行, 如果队列任务已满则阻塞等待直到有空位
func (g *gpool) Go(fn func() error, callback func(err error)) {
	job := g.newJob(fn, callback)
	select {
	case g.jobQueue <- job:
	case <-g.stop:
		callback(ErrGPoolClosed)
	}
}

// 同步执行
func (g *gpool) GoSync(fn func() error) (result error) {
	var wg sync.WaitGroup
	wg.Add(1)
	g.Go(fn, func(err error) {
		result = err
		wg.Done()
	})
	wg.Wait()
	return result
}

// 尝试异步执行, 如果任务队列已满则返回false
func (g *gpool) TryGo(fn func() error, callback func(err error)) (ok bool) {
	job := g.newJob(fn, callback)
	select {
	case g.jobQueue <- job:
		return true
	default:
		g.wg.Done()
		return false
	}
}

// 尝试同步执行, 如果任务队列已满则返回false
func (g *gpool) TryGoSync(fn func() error) (result error, ok bool) {
	var wg sync.WaitGroup
	wg.Add(1)
	ok = g.TryGo(fn, func(err error) {
		result = err
		wg.Done()
	})
	if ok {
		wg.Wait()
	}
	return
}

// 执行等待所有函数完成
func (g *gpool) GoAndWait(fn ...func() error) error {
	if len(fn) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(fn))

	for _, f := range fn {
		wg.Add(1)
		g.Go(f, func(err error) {
			if err != nil {
				errChan <- err
			}
			wg.Done()
		})
	}
	wg.Wait()

	var err error
	select {
	case err = <-errChan:
	default:
	}
	return err
}

// 等待队列中所有的任务结束
func (g *gpool) Wait() {
	g.wg.Wait()
}

// 关闭
//
// 命令所有没有收到任务的工人立即停工, 收到任务的工人完成当前任务后停工, 不管任务队列是否清空.
// 表现为加入队列的任务不一定会执行, 但是正在执行的任务不会被取消并会等待这些任务执行完毕.
func (g *gpool) Close() {
	g.onceClose.Do(func() {
		g.stop <- struct{}{}
		<-g.stop
		close(g.stop)
	})
}

func (g *gpool) newJob(fn func() error, callback func(err error)) *job {
	g.wg.Add(1)
	return newJob(fn, func(err error) {
		g.wg.Done()
		if callback != nil {
			callback(err)
		}
	})
}
