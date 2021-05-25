/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/3/19
   Description :
-------------------------------------------------
*/

package core

type IGPool interface {
	// 获取gpool
	GetGPoolGroup(name ...string) IGPoolGroup
	// 关闭
	Close()
}

type IGPoolGroup interface {
	// 异步执行
	Go(fn func() error) chan error
	// 同步执行
	GoSync(fn func() error) error
	// 关闭
	Close()
}
