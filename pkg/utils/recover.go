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
	PC       uintptr
	Function string
	File     string
	Line     int
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
	const depth = 16
	const defSkip = 4 // 调试结果
	var pcs [depth]uintptr
	n := runtime.Callers(defSkip, pcs[:])
	ff := runtime.CallersFrames(pcs[:n])
	for {
		f, ok := ff.Next()
		if !ok {
			break
		}
		callers = append(callers, &Caller{
			PC:       f.PC,
			Function: f.Function,
			File:     f.File,
			Line:     f.Line,
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
		callers[i+1] = fmt.Sprintf("%s:%d  %s", c.File, c.Line, c.Function)
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
		callers[i+1] = fmt.Sprintf("%s:%d  %s", c.File, c.Line, c.Function)
	}
	return strings.Join(callers, "\n")
}
