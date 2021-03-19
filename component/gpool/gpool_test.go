/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/3/19
   Description :
-------------------------------------------------
*/

package gpool

import (
	"testing"
	"time"
)

func TestGPool(t *testing.T) {
	g := newGPool(&GPoolConfig{
		JobQueueSize: 0,
		ThreadCount:  0,
	})
	defer g.Close()

	chs := make([]chan error, 10)
	for i := 0; i < len(chs); i++ {
		ch := g.Go(func(i int) func() error {
			return func() error {
				t.Log("Go", i)
				return nil
			}
		}(i))
		chs[i] = ch
	}
	for i, ch := range chs {
		if err := <-ch; err != nil {
			t.Error(i, err)
		}
	}

	for i := 0; i < 10; i++ {
		err := g.GoSync(func(i int) func() error {
			return func() error {
				t.Log("GoSync", i)
				return nil
			}
		}(i))

		if err != nil {
			t.Error(err)
		}
	}

	for i := 0; i < 10; i++ {
		go func(i int) {
			err := g.GoSync(func() error {
				t.Log("GoSync2", i)
				return nil
			})

			if err != nil {
				t.Error(err)
			}
		}(i)
	}
	time.Sleep(time.Second)
}
