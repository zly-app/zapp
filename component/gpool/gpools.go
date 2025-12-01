/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/3/19
   Description :
-------------------------------------------------
*/

package gpool

import (
	"go.uber.org/zap"

	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/config"
	"github.com/zly-app/zapp/consts"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/handler"
	"github.com/zly-app/zapp/log"
	"github.com/zly-app/zapp/pkg/utils"
)

type gpools struct {
	conn    *conn.AnyConn[core.IGPool]
	defPool core.IGPool
}

func NewCreator() core.IGPools {
	g := &gpools{
		defPool: NewGPool(new(GPoolConfig)),
	}
	g.conn = conn.NewAnyConn[core.IGPool](g.Close)
	handler.AddHandler(handler.AfterCloseComponent, func(_ core.IApp, _ handler.HandlerType) {
		g.conn.CloseAll()
	})
	return g
}

func (g *gpools) GetGPool(name ...string) core.IGPool {
	pool, err := g.conn.GetConn(g.makeGPoolGroup, name...)
	if err != nil {
		log.Warn("GetGPool call GetConn fail. use default pool", zap.Strings("name", name), zap.Error(err))
		return g.defPool
	}
	return pool
}

func (g *gpools) makeGPoolGroup(name string) (core.IGPool, error) {
	componentName := utils.Ternary.Or(name, consts.DefaultComponentName).(string)

	conf := new(GPoolConfig)
	err := config.Conf.ParseComponentConfig(DefaultComponentType, componentName, conf, true)
	if err != nil {
		log.Warn("makeGPoolGroup call ParseComponentConfig fail. use default config", zap.String("name", componentName), zap.Error(err))
	}
	conf.check()

	return NewGPool(conf), nil
}

func (g *gpools) Close(name string, pool core.IGPool) {
	pool.Close()
}
