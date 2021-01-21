/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2020/11/22
   Description :
-------------------------------------------------
*/

package utils

import (
	"errors"
	"fmt"
)

var Recover = new(recoverCli)

type recoverCli struct{}

func (*recoverCli) WrapCall(fn func() error) (err error) {
	// 包装执行, 拦截panic
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		switch v := e.(type) {
		case error:
			err = v
		case string:
			err = errors.New(v)
		default:
			err = errors.New(fmt.Sprint(e))
		}
	}()

	err = fn()
	return
}
