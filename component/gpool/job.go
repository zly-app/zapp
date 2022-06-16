/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/3/19
   Description :
-------------------------------------------------
*/

package gpool

import (
	"github.com/zly-app/zapp/pkg/utils"
)

type job struct {
	// 执行函数
	fn       func() error
	callback func(err error)
	err      error
}

func newJob(fn func() error, callback func(err error)) *job {
	j := &job{
		fn:       fn,
		callback: callback,
	}
	return j
}

// 执行
func (j *job) Do() {
	if j.fn != nil {
		j.err = utils.Recover.WrapCall(j.fn)
	}
	if j.callback != nil {
		j.callback(j.err)
	}
}
