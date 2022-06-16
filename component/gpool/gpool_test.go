/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/3/19
   Description :
-------------------------------------------------
*/

package gpool

import (
	"sync"
	"testing"
	"time"
)

func TestGo(t *testing.T) {
	g := NewGPool(new(GPoolConfig))
	defer g.Close()

	results := make([]error, 5)
	var wg sync.WaitGroup
	wg.Add(len(results))
	for i := 0; i < len(results); i++ {
		g.Go(func(i int) func() error {
			return func() error {
				time.Sleep(time.Second)
				t.Log("GoAsync", i)
				return nil
			}
		}(i), func(i int) func(err error) {
			return func(err error) {
				results[i] = err
				wg.Done()
			}
		}(i))
	}
	wg.Wait()
}

func TestGoSync(t *testing.T) {
	g := NewGPool(new(GPoolConfig))
	defer g.Close()

	for i := 0; i < 5; i++ {
		_ = g.GoSync(func(i int) func() error {
			return func() error {
				t.Log("GoSync", i)
				return nil
			}
		}(i))
	}

	var wg sync.WaitGroup
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func(i int) {
			_ = g.GoSync(func() error {
				t.Log("Async(GoSync)", i)
				wg.Done()
				return nil
			})
		}(i)
	}
	wg.Wait()
}

func TestTryGo(t *testing.T) {
	g := NewGPool(&GPoolConfig{
		ThreadCount: 1,
	})
	defer g.Close()

	ok := g.TryGo(func() error {
		time.Sleep(time.Millisecond * 200)
		t.Log("TryGo 1")
		return nil
	}, nil)
	if !ok {
		t.Error("TryGo 1 False")
	}

	ok = g.TryGo(func() error {
		time.Sleep(time.Millisecond * 200)
		t.Log("TryGo 2")
		return nil
	}, nil)
	if !ok {
		t.Error("TryGo 2 Ok")
	}
}

func TestTryGoSync(t *testing.T) {
	g := NewGPool(&GPoolConfig{
		ThreadCount: 1,
	})
	defer g.Close()

}

func TestClose(t *testing.T) {
	g := NewGPool(new(GPoolConfig))
	time.Sleep(time.Millisecond * 200)
	g.Close() // 此时已经获取到了 worker, 会走到获取 job 下的stop那里

	g = NewGPool(&GPoolConfig{
		ThreadCount: 1,
	})
	g.Go(func() error {
		time.Sleep(time.Second)
		return nil
	}, nil)
	g.Close() // 此时worker已被占用, 会走到获取 worker 下的stop那里
}

func TestWait(t *testing.T) {
	g := NewGPool(&GPoolConfig{
		ThreadCount: 1,
	})
	defer g.Close()
	g.Go(func() error {
		time.Sleep(time.Second)
		t.Log("run 1")
		return nil
	}, nil)
	g.Go(func() error {
		time.Sleep(time.Second)
		t.Log("run 2")
		return nil
	}, nil)
	g.Wait()
}
