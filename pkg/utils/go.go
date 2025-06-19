package utils

import (
	"sync"
)

var Go = goCli{}

type goCli struct{}

// 执行等待所有函数完成, 会自动 Recover, 如果有函数执行错误, 会返回第一个不为nil的error
func (goCli) GoAndWait(fns ...func() error) error {
	if len(fns) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(fns))

	for _, fn := range fns {
		wg.Add(1)
		go func(fn func() error) {
			defer wg.Done()
			if e := Recover.WrapCall(fn); e != nil {
				errChan <- e
			}
		}(fn)
	}
	wg.Wait()

	var err error
	select {
	case err = <-errChan:
	default:
	}
	return err
}

// 启用协程运行函数, 并返回一个wait函数等待所有函数执行完成, 会自动 Recover, 如果有函数执行错误, 会返回第一个不为nil的error
func (goCli) GoRetWait(fns ...func() error) func() error {
	if len(fns) == 0 {
		return func() error { return nil }
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(fns))

	for _, fn := range fns {
		wg.Add(1)
		go func(fn func() error) {
			defer wg.Done()
			if e := Recover.WrapCall(fn); e != nil {
				errChan <- e
			}
		}(fn)
	}

	return func() error {
		wg.Wait()

		var err error
		select {
		case err = <-errChan:
		default:
		}
		return err
	}
}
