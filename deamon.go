/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/2/19
   Description :
-------------------------------------------------
*/

package zapp

import (
	"flag"
	"os"

	"github.com/takama/daemon"
	"go.uber.org/zap"

	"github.com/zly-app/zapp/logger"
)

func (app *appCli) enableDaemon() {
	if !app.opt.EnableDaemon || len(os.Args) < 2 {
		return
	}

	flag.String("install", "", "安装服务, string 是运行时传递给 app 的参数, 请不要使用相对路径")
	flag.Bool("remove", false, "移除服务")
	flag.Bool("start", false, "启动服务")
	flag.Bool("stop", false, "停止服务")
	flag.Bool("status", false, "查看运行状态")

	switch os.Args[1] {
	case "install":
	case "remove":
	case "start":
	case "stop":
	case "status":
	default:
		return
	}

	d, err := daemon.New(app.name, app.name, daemon.SystemDaemon)
	if err != nil {
		logger.Log.Fatal("守护进程模块创建失败", zap.Error(err))
	}

	var out string
	switch os.Args[1] {
	case "install":
		out, err = d.Install(os.Args[2:]...)
	case "remove":
		out, err = d.Remove()
	case "start":
		out, err = d.Start()
	case "stop":
		out, err = d.Stop()
	case "status":
		out, err = d.Status()
	}

	if err != nil {
		logger.Log.Error(out, zap.Error(err))
		os.Exit(1)
	}

	logger.Log.Info(out)
	os.Exit(0)
}
