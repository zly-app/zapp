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
	g := NewGPool(new(GPoolConfig))
	defer g.Close()

	chs := make([]<-chan error, 5)
	for i := 0; i < len(chs); i++ {
		ch := g.Go(func(i int) func() error {
			return func() error {
				time.Sleep(time.Second)
				t.Log("Go", i)
				return nil
			}
		}(i))
		chs[i] = ch
	}
	for _, ch := range chs {
		if ch == nil {
			continue
		}
		<-ch
	}

	for i := 0; i < 5; i++ {
		_ = g.GoSync(func(i int) func() error {
			return func() error {
				t.Log("GoSync", i)
				return nil
			}
		}(i))
	}

	for i := 0; i < 5; i++ {
		go func(i int) {
			_ = g.GoSync(func() error {
				t.Log("GoSync2", i)
				return nil
			})
		}(i)
	}
	time.Sleep(time.Second)
}
