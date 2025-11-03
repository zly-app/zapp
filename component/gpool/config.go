/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/3/19
   Description :
-------------------------------------------------
*/

package gpool

import (
	"github.com/zly-app/zapp/core"
)

// 默认组件类型
const DefaultComponentType core.ComponentType = "gpool"

const (
	// 最小任务队列大小
	defaultMinJobQueueSize = 100000
	// 最小线程数
	defaultMinThreadCount = 100
)

type GPoolConfig struct {
	// 任务队列大小
	JobQueueSize int
	// 同时处理信息的goroutine数, 设为0时取逻辑cpu数量 * 2, 设为负数时不作任何限制, 每个请求有独立的线程执行
	ThreadCount int
}

func (g *GPoolConfig) check() {
	if g.JobQueueSize < defaultMinJobQueueSize {
		g.JobQueueSize = defaultMinJobQueueSize
	}
	if g.ThreadCount < defaultMinThreadCount {
		g.ThreadCount = defaultMinThreadCount
	}
	if g.ThreadCount < 0 {
		g.ThreadCount = -1
	}
}
