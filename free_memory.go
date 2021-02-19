/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/2/19
   Description :
-------------------------------------------------
*/

package zapp

import (
	"runtime/debug"
	"time"
)

// 开始释放内存
func (app *appCli) startFreeMemory() {
	go app.freeMemory()
}

func (app *appCli) freeMemory() {
	interval := app.config.Config().Frame.FreeMemoryInterval
	if interval <= 0 {
		return
	}

	t := time.NewTicker(time.Duration(interval) * time.Millisecond)
	for {
		select {
		case <-app.baseCtx.Done():
			t.Stop()
			return
		case <-t.C:
			debug.FreeOSMemory()
		}
	}
}
