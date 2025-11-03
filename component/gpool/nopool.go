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
	defer n.wg.Done()
	err := fn()
	if callback != nil {
		callback(err)
	}
}

func (n *NoPool) GoSync(fn func() error) (result error) {
	n.wg.Add(1)
	defer n.wg.Done()
	return fn()
}

func (n *NoPool) TryGo(fn func() error, callback func(err error)) (ok bool) {
	n.wg.Add(1)
	defer n.wg.Done()
	err := fn()
	if callback != nil {
		callback(err)
	}
	return true
}

func (n *NoPool) TryGoSync(fn func() error) (result error, ok bool) {
	n.wg.Add(1)
	defer n.wg.Done()
	return fn(), true
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
