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
	"strings"
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

func (*recoverCli) IsRecoverError(err error) bool {
	_, ok := err.(RecoverError)
	return ok
}

// 包装函数, 捕获panic
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

// 获取 recover 错误
func (*recoverCli) GetRecoverError(err error) (RecoverError, bool) {
	re, ok := err.(RecoverError)
	return re, ok
}

// 获取 recover 错误的详情
func (*recoverCli) GetRecoverErrors(err error) []string {
	re, ok := err.(RecoverError)
	if !ok {
		return []string{err.Error()}
	}

	var callers []string
	callers = make([]string, len(re.Callers())+1)
	callers[0] = err.Error()
	for i, c := range re.Callers() {
		callers[i+1] = fmt.Sprintf("%s:%d", c.File, c.Line)
	}
	return callers
}

// 获取 recover 错误的详情
func (*recoverCli) GetRecoverErrorDetail(err error) string {
	re, ok := err.(RecoverError)
	if !ok {
		return err.Error()
	}

	var callers []string
	callers = make([]string, len(re.Callers())+1)
	callers[0] = err.Error()
	for i, c := range re.Callers() {
		callers[i+1] = fmt.Sprintf("%s:%d", c.File, c.Line)
	}
	return strings.Join(callers, "\n")
}
