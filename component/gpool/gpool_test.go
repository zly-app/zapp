/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/3/19
   Description :
-------------------------------------------------
*/

package gpool

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGo(t *testing.T) {
	g := NewGPool(new(GPoolConfig))
	defer g.Close()

	var count int32
	var wg sync.WaitGroup
	wg.Add(5)
	for i := 0; i < 5; i++ {
		g.Go(func() error {
			atomic.AddInt32(&count, 1)
			return nil
		}, func(err error) {
			wg.Done()
		})
	}
	wg.Wait()
	require.Equal(t, int32(5), atomic.LoadInt32(&count))
}

func TestGoSync(t *testing.T) {
	g := NewGPool(new(GPoolConfig))
	defer g.Close()

	err := g.GoSync(func() error {
		return nil
	})
	require.Nil(t, err)

	err = g.GoSync(func() error {
		return errors.New("some error")
	})
	require.Equal(t, "some error", err.Error())
}

func TestTryGo(t *testing.T) {
	g := NewGPool(new(GPoolConfig))
	defer g.Close()

	// 正常情况下 TryGo 应成功
	var count int32
	var wg sync.WaitGroup
	wg.Add(3)
	for i := 0; i < 3; i++ {
		ok := g.TryGo(func() error {
			atomic.AddInt32(&count, 1)
			return nil
		}, func(err error) {
			wg.Done()
		})
		require.True(t, ok, "TryGo should succeed when pool has capacity")
	}
	wg.Wait()
	require.Equal(t, int32(3), atomic.LoadInt32(&count))
}

func TestTryGoSync(t *testing.T) {
	g := NewGPool(new(GPoolConfig))
	defer g.Close()

	result, ok := g.TryGoSync(func() error {
		return errors.New("test")
	})
	require.True(t, ok)
	require.Equal(t, "test", result.Error())
}

func TestClose(t *testing.T) {
	// 场景1: Close 后 Go 应返回 ErrGPoolClosed
	g := NewGPool(&GPoolConfig{
		ThreadCount: 1,
	})
	g.Close()

	var receivedErr error
	var wg sync.WaitGroup
	wg.Add(1)
	g.Go(func() error {
		return nil
	}, func(err error) {
		receivedErr = err
		wg.Done()
	})
	wg.Wait()
	require.Equal(t, ErrGPoolClosed, receivedErr, "Go after Close should receive ErrGPoolClosed")

	// 场景2: 正在执行的任务应完成
	g2 := NewGPool(&GPoolConfig{
		ThreadCount: 1,
	})
	var executed int32
	var jobStarted int32
	g2.Go(func() error {
		atomic.StoreInt32(&jobStarted, 1)
		time.Sleep(time.Millisecond * 100)
		atomic.StoreInt32(&executed, 1)
		return nil
	}, nil)
	// 等待 job 确实被 worker 开始执行
	for atomic.LoadInt32(&jobStarted) == 0 {
		time.Sleep(time.Millisecond * 5)
	}
	g2.Close()
	require.Equal(t, int32(1), atomic.LoadInt32(&executed), "running job should complete before Close returns")

	// 场景3: Close 后队列中 job 的 callback 一定会被调用（可能收到 ErrGPoolClosed 或正常执行完）
	g3 := NewGPool(&GPoolConfig{
		ThreadCount:  1,
		JobQueueSize: 100000,
	})
	var callbackCalled int32
	var callbackWg sync.WaitGroup
	// 占住 worker
	var blockerStarted3 int32
	g3.Go(func() error {
		atomic.StoreInt32(&blockerStarted3, 1)
		time.Sleep(time.Millisecond * 500)
		return nil
	}, nil)
	for atomic.LoadInt32(&blockerStarted3) == 0 {
		time.Sleep(time.Millisecond * 5)
	}
	// 放入队列的任务
	callbackWg.Add(1)
	g3.Go(func() error {
		return nil
	}, func(err error) {
		atomic.StoreInt32(&callbackCalled, 1)
		callbackWg.Done()
	})
	g3.Close()
	callbackWg.Wait()
	require.Equal(t, int32(1), atomic.LoadInt32(&callbackCalled), "queued job callback must be called even after Close")

	// 场景4: 重复 Close 不应 panic
	g4 := NewGPool(&GPoolConfig{
		ThreadCount: 1,
	})
	g4.Close()
	g4.Close()
}

func TestGoAndWait(t *testing.T) {
	g := NewGPool(new(GPoolConfig))
	defer g.Close()

	err := g.GoAndWait()
	require.Nil(t, err)

	err = g.GoAndWait(func() error {
		return nil
	})
	require.Nil(t, err)

	err = g.GoAndWait(func() error {
		return nil
	}, func() error {
		return errors.New("2")
	}, func() error {
		return errors.New("3")
	})
	require.NotNil(t, err)
}

func TestWait(t *testing.T) {
	g := NewGPool(&GPoolConfig{
		ThreadCount: 1,
	})
	defer g.Close()
	var count int32
	g.Go(func() error {
		atomic.AddInt32(&count, 1)
		return nil
	}, nil)
	g.Go(func() error {
		atomic.AddInt32(&count, 1)
		return nil
	}, nil)
	g.Wait()
	require.Equal(t, int32(2), atomic.LoadInt32(&count))
}

func TestGoRetWait(t *testing.T) {
	g := NewGPool(new(GPoolConfig))
	defer g.Close()

	// 空调用
	waitFn := g.GoRetWait()
	require.Nil(t, waitFn())

	// 正常执行
	waitFn = g.GoRetWait(
		func() error { return nil },
		func() error { return errors.New("err") },
	)
	err := waitFn()
	require.NotNil(t, err)
}
