package gpool

import (
	"sync"

	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/pkg/utils"
)

type NoPool struct {
	wg sync.WaitGroup
}

func (n *NoPool) Go(fn func() error, callback func(err error)) {
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		err := utils.Recover.WrapCall(fn)
		if callback != nil {
			callback(err)
		}
	}()
}

func (n *NoPool) GoSync(fn func() error) (result error) {
	var wg sync.WaitGroup
	wg.Add(1)
	n.Go(fn, func(err error) {
		result = err
		wg.Done()
	})
	wg.Wait()
	return result
}

func (n *NoPool) TryGo(fn func() error, callback func(err error)) (ok bool) {
	n.Go(fn, callback)
	return true
}

func (n *NoPool) TryGoSync(fn func() error) (result error, ok bool) {
	result = n.GoSync(fn)
	return result, true
}

func (n *NoPool) GoAndWait(fn ...func() error) error {
	return utils.Go.GoAndWait(fn...)
}

func (n *NoPool) GoRetWait(fn ...func() error) func() error { return utils.Go.GoRetWait(fn...) }

func (n *NoPool) Wait() {
	n.wg.Wait()
}

func (n *NoPool) Close() {}

func NewNoPool() core.IGPool {
	return &NoPool{}
}
