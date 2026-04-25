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
	"runtime"
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

// === 新增测试 ===

func TestConfigCheck_ThreadCountZero(t *testing.T) {
	// ThreadCount=0 应该被设置为 NumCPU*2
	conf := &GPoolConfig{ThreadCount: 0}
	conf.check()
	expected := runtime.NumCPU() * 2
	if expected < defaultMinThreadCount {
		expected = defaultMinThreadCount
	}
	require.Equal(t, expected, conf.ThreadCount, "ThreadCount=0 should be set to NumCPU*2 (or min)")
}

func TestConfigCheck_ThreadCountNegative(t *testing.T) {
	conf := &GPoolConfig{ThreadCount: -5}
	conf.check()
	require.Equal(t, -1, conf.ThreadCount, "negative ThreadCount should be normalized to -1")
}

func TestConfigCheck_ThreadCountSmall(t *testing.T) {
	conf := &GPoolConfig{ThreadCount: 10}
	conf.check()
	require.Equal(t, defaultMinThreadCount, conf.ThreadCount, "ThreadCount below minimum should be raised")
}

func TestConfigCheck_JobQueueSize(t *testing.T) {
	conf := &GPoolConfig{JobQueueSize: 50}
	conf.check()
	require.Equal(t, defaultMinJobQueueSize, conf.JobQueueSize, "JobQueueSize below minimum should be raised")
}

func TestNoPool_Go(t *testing.T) {
	n := NewNoPool()

	var count int32
	var wg sync.WaitGroup
	wg.Add(3)
	for i := 0; i < 3; i++ {
		n.Go(func() error {
			atomic.AddInt32(&count, 1)
			return nil
		}, func(err error) {
			wg.Done()
		})
	}
	wg.Wait()
	require.Equal(t, int32(3), atomic.LoadInt32(&count), "NoPool.Go should execute asynchronously")
}

func TestNoPool_GoSync(t *testing.T) {
	n := NewNoPool()

	err := n.GoSync(func() error {
		return errors.New("nopool sync error")
	})
	require.Equal(t, "nopool sync error", err.Error())
}

func TestNoPool_TryGo(t *testing.T) {
	n := NewNoPool()

	var count int32
	var wg sync.WaitGroup
	wg.Add(1)
	ok := n.TryGo(func() error {
		atomic.AddInt32(&count, 1)
		return nil
	}, func(err error) {
		wg.Done()
	})
	require.True(t, ok, "NoPool.TryGo should always return true")
	wg.Wait()
	require.Equal(t, int32(1), atomic.LoadInt32(&count))
}

func TestNoPool_TryGoSync(t *testing.T) {
	n := NewNoPool()

	result, ok := n.TryGoSync(func() error {
		return errors.New("nopool try sync error")
	})
	require.True(t, ok)
	require.Equal(t, "nopool try sync error", result.Error())
}

func TestNoPool_GoAndWait(t *testing.T) {
	n := NewNoPool()

	err := n.GoAndWait(func() error {
		return nil
	}, func() error {
		return errors.New("err")
	})
	require.NotNil(t, err)
}

func TestNoPool_GoRetWait(t *testing.T) {
	n := NewNoPool()

	waitFn := n.GoRetWait(func() error { return nil })
	require.Nil(t, waitFn())

	waitFn = n.GoRetWait(func() error { return errors.New("ret err") })
	require.NotNil(t, waitFn())
}

func TestNoPool_Wait(t *testing.T) {
	n := NewNoPool()
	var executed int32
	n.Go(func() error {
		atomic.StoreInt32(&executed, 1)
		return nil
	}, nil)
	n.Wait()
	require.Equal(t, int32(1), atomic.LoadInt32(&executed), "NoPool.Wait should wait for Go tasks")
}

func TestNoPool_Close(t *testing.T) {
	n := NewNoPool()
	// Close 应该是空操作，不 panic
	n.Close()
	n.Close()
}

func TestNoPool_PanicRecovery(t *testing.T) {
	n := NewNoPool()

	var callbackErr error
	var wg sync.WaitGroup
	wg.Add(1)
	n.Go(func() error {
		panic("test panic")
	}, func(err error) {
		callbackErr = err
		wg.Done()
	})
	wg.Wait()
	require.NotNil(t, callbackErr, "NoPool.Go should recover panic and pass error to callback")
}

func TestGPool_PanicRecovery(t *testing.T) {
	g := NewGPool(&GPoolConfig{ThreadCount: 1})
	defer g.Close()

	var callbackErr error
	var wg sync.WaitGroup
	wg.Add(1)
	g.Go(func() error {
		panic("gpool panic")
	}, func(err error) {
		callbackErr = err
		wg.Done()
	})
	wg.Wait()
	require.NotNil(t, callbackErr, "gpool.Go should recover panic and pass error to callback")
}

func TestGPool_GoSyncAfterClose(t *testing.T) {
	g := NewGPool(&GPoolConfig{ThreadCount: 1})
	g.Close()

	err := g.GoSync(func() error {
		return nil
	})
	require.Equal(t, ErrGPoolClosed, err, "GoSync after Close should return ErrGPoolClosed")
}

func TestGPool_ConcurrentGo(t *testing.T) {
	g := NewGPool(&GPoolConfig{ThreadCount: 4})
	defer g.Close()

	var count int32
	const n = 100
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		g.Go(func() error {
			atomic.AddInt32(&count, 1)
			return nil
		}, func(err error) {
			wg.Done()
		})
	}
	wg.Wait()
	require.Equal(t, int32(n), atomic.LoadInt32(&count))
}

func TestNewGPool_NegativeThreadCount(t *testing.T) {
	// ThreadCount < 0 应该返回 NoPool
	g := NewGPool(&GPoolConfig{ThreadCount: -1})
	_, ok := g.(*NoPool)
	require.True(t, ok, "negative ThreadCount should return NoPool")
}

func TestGPool_CallbackNil(t *testing.T) {
	g := NewGPool(&GPoolConfig{ThreadCount: 1})
	defer g.Close()

	// callback 为 nil 不应 panic
	g.Go(func() error {
		return nil
	}, nil)
	g.Wait()
}

func TestGPool_GoSyncError(t *testing.T) {
	g := NewGPool(&GPoolConfig{ThreadCount: 1})
	defer g.Close()

	err := g.GoSync(func() error {
		return errors.New("sync error")
	})
	require.Equal(t, "sync error", err.Error())
}
