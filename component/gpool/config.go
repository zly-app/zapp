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
	// 默认最小任务队列大小
	defaultMinJobQueueSize = 100
)

type GPoolConfig struct {
	// 任务队列大小
	JobQueueSize int
	// 同时处理信息的goroutine数, 默认为逻辑cpu数量
	ThreadCount int
}

func (g *GPoolConfig) check() {
	if g.JobQueueSize <= 0 {
		g.JobQueueSize = defaultJobQueueSize
	}
	if g.JobQueueSize < defaultMinJobQueueSize {
		g.JobQueueSize = defaultMinJobQueueSize
	}
	if g.ThreadCount <= 0 {
		g.ThreadCount = runtime.NumCPU()
	}
}
