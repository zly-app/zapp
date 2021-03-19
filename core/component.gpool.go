/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/3/19
   Description :
-------------------------------------------------
*/

package core

type IGPoolManager interface {
	// 获取gpool
	GetGPool(name ...string) IGPool
	// 关闭
	Close()
}

type IGPool interface {
	// 异步执行
	Go(fn func() error) chan error
	// 同步执行
	GoSync(fn func() error) error
	// 关闭
	Close()
}
