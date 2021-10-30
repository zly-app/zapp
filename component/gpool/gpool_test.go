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

	"github.com/zly-app/zapp/core"
)

func TestGPool(t *testing.T) {
	g := NewGPool(new(GPoolConfig))
	defer g.Close()

	results := make([]core.IGPoolJobResult, 5)
	for i := 0; i < len(results); i++ {
		result := g.Go(func(i int) func() error {
			return func() error {
				time.Sleep(time.Second)
				t.Log("GoAsync", i)
				return nil
			}
		}(i))
		results[i] = result
	}
	for _, result := range results {
		_ = result.Wait()
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
				t.Log("Async(GoSync)", i)
				return nil
			})
		}(i)
	}
	time.Sleep(time.Second)
}
