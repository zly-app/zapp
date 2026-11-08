package freemem

import (
	"context"
	"runtime/debug"
	"time"
)

// Releaser 是内存释放模型：拥有「释放间隔」这一份配置数据，以及
// 「按间隔调用 debug.FreeOSMemory，上下文结束时停表」这条规则。
//
// 允许调用：context / runtime/debug（底座机制）。禁止调用任何上层模型。
type Releaser struct {
	interval time.Duration
}

// New 按毫秒间隔创建释放器；intervalMS <= 0 表示禁用（表现为空操作）
func New(intervalMS int) *Releaser {
	if intervalMS <= 0 {
		return &Releaser{}
	}
	return &Releaser{interval: time.Duration(intervalMS) * time.Millisecond}
}

// Enabled 报告是否启用
func (r *Releaser) Enabled() bool {
	return r.interval > 0
}

// Start 在后台按间隔释放内存，ctx 取消时退出
//
// 未启用时直接返回，不启动协程 —— 没有终止条件的常驻协程是不允许的。
func (r *Releaser) Start(ctx context.Context) {
	if !r.Enabled() {
		return
	}
	go func() {
		t := time.NewTicker(r.interval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				debug.FreeOSMemory()
			}
		}
	}()
}
