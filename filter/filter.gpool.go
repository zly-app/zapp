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

type GPoolFilterConfig struct {
	Config  *gpool.GPoolConfig
	Client  map[string]map[string]*gpool.GPoolConfig
	Service map[string]*gpool.GPoolConfig
}
type gPoolFilter struct {
	once sync.Once

	defPool core.IGPool

	clientPool  map[string]map[string]core.IGPool
	servicePool map[string]core.IGPool
}

func (g *gPoolFilter) Init(app core.IApp) error {
	var err error
	g.once.Do(func() {
		conf := &GPoolFilterConfig{}
		err = config.Conf.ParseFilterConfig("base.gpool", conf, true)
		if err != nil {
			return
		}

		g.defPool = gpool.NewGPool(g.genGPoolConfig(conf.Config))

		g.clientPool = make(map[string]map[string]core.IGPool, len(conf.Client))
		for clientType, clientConf := range conf.Client {
			chain, ok := g.clientPool[clientType]
			if !ok {
				chain = make(map[string]core.IGPool)
				g.clientPool[clientType] = chain
			}

			for clientName, gConf := range clientConf {
				chain[clientName] = gpool.NewGPool(g.genGPoolConfig(gConf))
			}
		}

		g.servicePool = make(map[string]core.IGPool, len(conf.Service))
		for name, gConf := range conf.Service {
			g.servicePool[name] = gpool.NewGPool(g.genGPoolConfig(gConf))
		}
	})
	return err
}
func (g *gPoolFilter) genGPoolConfig(raw *gpool.GPoolConfig) *gpool.GPoolConfig {
	if raw == nil {
		raw = new(gpool.GPoolConfig)
	}
	return raw
}

func (g *gPoolFilter) HandleInject(ctx context.Context, req, rsp interface{}, next core.FilterInjectFunc) error {
	utils.Otel.CtxEvent(ctx, "gpool.wait")
	err := g.getGPool(ctx).GoSync(func() error {
		utils.Otel.CtxEvent(ctx, "gpool.do")
		return next(ctx, req, rsp)
	})
	return err
}

func (g *gPoolFilter) Handle(ctx context.Context, req interface{}, next core.FilterFunc) (interface{}, error) {
	utils.Otel.CtxEvent(ctx, "gpool.wait")
	var rsp interface{}
	err := g.getGPool(ctx).GoSync(func() error {
		utils.Otel.CtxEvent(ctx, "gpool.do")
		var e error
		rsp, e = next(ctx, req)
		return e
	})
	return rsp, err
}

func (g *gPoolFilter) getClientGPool(clientType, clientName string) core.IGPool {
	ct, ok := g.clientPool[clientType]
	if ok {
		gp, ok := ct[clientName]
		if ok {
			return gp
		}
		gp, ok = ct[defName] // 默认客户端组件
		if ok {
			return gp
		}
	}

	ct, ok = g.clientPool[defName]
	if ok {
		gp, ok := ct[defName]
		if ok {
			return gp
		}
	}
	return g.defPool
}
func (g *gPoolFilter) getServicGPool(serviceName string) core.IGPool {
	gp, ok := g.servicePool[serviceName]
	if ok {
		return gp
	}
	gp, ok = g.servicePool[defName]
	if ok {
		return gp
	}
	return g.defPool
}
func (g *gPoolFilter) getGPool(ctx context.Context) core.IGPool {
	meta := GetCallMeta(ctx)
	if meta.IsClientMeta() {
		return g.getClientGPool(meta.ClientType(), meta.ClientName())
	} else if meta.IsServiceMeta() {
		return g.getServicGPool(meta.ServiceName())
	}
	return g.defPool
}

func (g *gPoolFilter) Close() error {
	g.defPool.Close()
	for _, clientPool := range g.clientPool {
		for _, client := range clientPool {
			client.Close()
		}
	}
	for _, servicePool := range g.servicePool {
		servicePool.Close()
	}
	return nil
}
