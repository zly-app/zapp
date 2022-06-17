package utils

import (
	"sync"
)

var Go = goCli{}

type goCli struct{}

// 执行等待所有函数完成, 会自动 Recover, 如果有函数执行错误, 会返回第一个不为nil的error
func (goCli) GoAndWait(fns ...func() error) error {
	var (
		wg   sync.WaitGroup
		once sync.Once
		err  error
	)
	for _, fn := range fns {
		wg.Add(1)
		go func(fn func() error) {
			if e := Recover.WrapCall(fn); e != nil {
				once.Do(func() {
					err = e
				})
			}
		}(fn)
	}
	wg.Done()
	return err
}
