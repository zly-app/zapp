package gpool

import (
	"github.com/zly-app/zapp/pkg/utils"
)

// 工人
type worker struct {
	pool       chan<- *worker
	jobChannel chan *job     // 工作任务
	stop       chan struct{} // 停止信号
}

// 准备好
func (w *worker) Ready() {
	go func() {
		var job *job
		for {
			select {
			case job = <-w.jobChannel: // 等待任务
				err := utils.Recover.WrapCall(job.fn)
				job.done <- err
				w.pool <- w
			case <-w.stop:
				w.stop <- struct{}{}
				return
			}
		}
	}()
}

// 做任务
func (w *worker) Do(job *job) {
	w.jobChannel <- job
}

// 停止
func (w *worker) Stop() {
	w.stop <- struct{}{}
	<-w.stop
}

// 创建一个工人
func newWorker(pool chan<- *worker) *worker {
	return &worker{
		pool:       pool,
		jobChannel: make(chan *job),
		stop:       make(chan struct{}),
	}
}
