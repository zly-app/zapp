/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/3/19
   Description :
-------------------------------------------------
*/

package core

type IGPools interface {
	// 获取gpool
	GetGPool(name ...string) IGPool
	// 关闭
	Close()
}

type IGPool interface {
	// 异步执行
	Go(fn func() error, callback func(err error))
	// 同步执行
	GoSync(fn func() error) (result error)
	// 尝试异步执行, 如果任务队列已满则返回false
	TryGo(fn func() error, callback func(err error)) (ok bool)
	// 尝试同步执行, 如果任务队列已满则返回false
	TryGoSync(fn func() error) (result error, ok bool)
	// 等待队列中所有的任务结束
	Wait()
	// 关闭, 命令所有没有收到任务的工人立即停工, 收到任务的工人完成当前任务后停工, 不管任务队列是否清空
	Close()
}
