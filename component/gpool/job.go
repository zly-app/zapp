/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/3/19
   Description :
-------------------------------------------------
*/

package gpool

import (
	"sync"

	"github.com/zly-app/zapp/pkg/utils"
)

type job struct {
	// 执行函数
	fn  func() error
	wg  sync.WaitGroup
	err error
}

func newJob(fn func() error) *job {
	j := &job{
		fn: fn,
	}
	j.wg.Add(1)
	return j
}

// 执行
func (j *job) Do() {
	j.err = utils.Recover.WrapCall(j.fn)
	j.wg.Done()
}

// 等待结果
func (j *job) Wait() error {
	j.wg.Wait()
	return j.err
}
