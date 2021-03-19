/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/3/19
   Description :
-------------------------------------------------
*/

package gpool

type job struct {
	// 执行函数
	fn func() error
	// 结束通知通道
	done chan error
}

func newJob(fn func() error) *job {
	return &job{
		fn:   fn,
		done: make(chan error, 1),
	}
}
