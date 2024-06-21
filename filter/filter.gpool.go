package filter

import (
	"context"
	"sync"

	"github.com/zly-app/zapp/component/gpool"
	"github.com/zly-app/zapp/config"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/pkg/utils"
)

func init() {
	RegisterFilterCreator("base.gpool", newGPoolFilter, newGPoolFilter)
}

var defGPoolFilter core.Filter

func newGPoolFilter() core.Filter {
	if defGPoolFilter == nil {
		defGPoolFilter = &gPoolFilter{}
	}
	return defGPoolFilter
}

type gPoolFilter struct {
	Config gpool.GPoolConfig
	once   sync.Once
	pool   core.IGPool
}

func (g *gPoolFilter) Init(app core.IApp) error {
	var err error
	g.once.Do(func() {
		err = config.Conf.ParseFilterConfig("base.gpool", g, true)
		if err != nil {
			return
		}
		g.pool = gpool.NewGPool(&g.Config)
	})
	return err
}

func (g *gPoolFilter) HandleInject(ctx context.Context, req, rsp interface{}, next core.FilterInjectFunc) error {
	utils.Otel.CtxEvent(ctx, "gpool.wait")
	err := g.pool.GoSync(func() error {
		utils.Otel.CtxEvent(ctx, "gpool.do")
		return next(ctx, req, rsp)
	})
	return err
}

func (g *gPoolFilter) Handle(ctx context.Context, req interface{}, next core.FilterFunc) (interface{}, error) {
	utils.Otel.CtxEvent(ctx, "gpool.wait")
	var rsp interface{}
	err := g.pool.GoSync(func() error {
		utils.Otel.CtxEvent(ctx, "gpool.do")
		var e error
		rsp, e = next(ctx, req)
		return e
	})
	return rsp, err
}

func (g *gPoolFilter) Close() error { return nil }
