package zapp

import (
	"github.com/zly-app/zapp/filter"
)

func (app *appCli) makeFilter() {
	app.Info("构建过滤器")
	app.handler(BeforeMakeFilter)
	filter.MakeFilter()
	filter.InitFilter(app)
	app.handler(AfterMakeFilter)
}

func (app *appCli) closeFilter() {
	app.Info("关闭过滤器")
	app.handler(BeforeCloseFilter)
	filter.CloseFilter()
	app.handler(AfterCloseFilter)
}
