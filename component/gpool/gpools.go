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
	"github.com/zly-app/zapp/logger"
	"github.com/zly-app/zapp/pkg/utils"
)

type gpools struct {
	conn *conn.Conn
}

func NewGPools() core.IGPools {
	return &gpools{
		conn: conn.NewConn(),
	}
}

func (g *gpools) GetGPool(name ...string) core.IGPool {
	return g.conn.GetInstance(g.makeGPoolGroup, name...).(core.IGPool)
}

func (g *gpools) makeGPoolGroup(name string) (conn.IInstance, error) {
	componentName := utils.Ternary.Or(name, consts.DefaultComponentName).(string)
	key := "components." + string(DefaultComponentType) + "." + componentName

	conf := new(GPoolConfig)

	vi := config.Conf.GetViper()
	if !vi.IsSet(key) {
		logger.Log.Warn("gpool组件配置不存在, 将使用默认配置", zap.String("name", componentName))
	} else if err := vi.UnmarshalKey(key, conf); err != nil {
		logger.Log.Warn("gpool组件配置解析失败, 将使用默认配置", zap.String("name", componentName), zap.Error(err))
	}

	conf.check()
	return newGPool(conf), nil
}

func (g *gpools) Close() {
	g.conn.CloseAll()
}
