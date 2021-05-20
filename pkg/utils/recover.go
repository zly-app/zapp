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
	"runtime"
)

type Caller struct {
	PC   uintptr
	File string
	Line int
}

type RecoverError interface {
	error
	Err() error
	Callers() []*Caller
}
type recoverError struct {
	err     error
	callers []*Caller
}

func (r recoverError) Error() string {
	return r.err.Error()
}
func (r recoverError) Err() error {
	return r.err
}
func (r recoverError) Callers() []*Caller {
	return r.callers
}
func newRecoverError(err error) RecoverError {
	var callers []*Caller
	for i := 3; ; i++ {
		pc, file, line, got := runtime.Caller(i)
		if !got {
			break
		}

		callers = append(callers, &Caller{
			PC:   pc,
			File: file,
			Line: line,
		})
	}

	return recoverError{
		err:     err,
		callers: callers,
	}
}

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
		err = newRecoverError(err)
	}()

	err = fn()
	return
}
