/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/3/19
   Description :
-------------------------------------------------
*/

package gpool

import (
	"runtime"

	"github.com/zly-app/zapp/core"
)

// 默认组件类型
const DefaultComponentType core.ComponentType = "gpool"

const (
	// 默认任务队列大小
	defaultJobQueueSize = 10000
)

type GPoolConfig struct {
	// 任务队列大小
	JobQueueSize int
	// 同时处理信息的goroutine数, 设为0时取逻辑cpu数量 * 2, 设为负数时不作任何限制, 每个请求有独立的线程执行
	ThreadCount int
}

func (g *GPoolConfig) check() {
	if g.JobQueueSize < 1 {
		g.JobQueueSize = defaultJobQueueSize
	}
	if g.ThreadCount == 0 {
		g.ThreadCount = runtime.NumCPU() * 2
	}
	if g.ThreadCount < 1 {
		g.ThreadCount = -1
	}
}
