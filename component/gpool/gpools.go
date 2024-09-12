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

	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/config"
	"github.com/zly-app/zapp/consts"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/handler"
	"github.com/zly-app/zapp/logger"
	"github.com/zly-app/zapp/pkg/utils"
)

type gpools struct {
	conn *conn.Conn
}

func NewCreator() core.IGPools {
	g := &gpools{
		conn: conn.NewConn(),
	}
	zapp.AddHandler(zapp.AfterCloseComponent, func(_ core.IApp, _ handler.HandlerType) {
		g.Close()
	})
	return g
}

func (g *gpools) GetGPool(name ...string) core.IGPool {
	return g.conn.GetInstance(g.makeGPoolGroup, name...).(core.IGPool)
}

func (g *gpools) makeGPoolGroup(name string) (conn.IInstance, error) {
	componentName := utils.Ternary.Or(name, consts.DefaultComponentName).(string)

	conf := new(GPoolConfig)
	err := config.Conf.ParseComponentConfig(DefaultComponentType, componentName, conf, true)
	if err != nil {
		logger.Log.Warn("gpool组件配置解析失败, 将使用默认配置", zap.String("name", componentName), zap.Error(err))
	}
	conf.check()

	return NewGPool(conf), nil
}

func (g *gpools) Close() {
	g.conn.CloseAll()
}
