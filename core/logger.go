/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/7/2
   Description :
-------------------------------------------------
*/

package core

import (
	"github.com/zlyuancn/zlog"
)

// 记录器
type ILogger interface {
	// 创建一个镜像log
	NewMirrorLogger(tag ...string) ILogger

	zlog.Loger
}
