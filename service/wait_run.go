/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/12/14
   Description :
-------------------------------------------------
*/

package service

import (
	"time"

	"go.uber.org/zap"

	"github.com/zly-app/zapp/core"
)

// 等待运行选项
type WaitRunOption struct {
	// 服务类型
	ServiceType core.ServiceType
	// 如果错误是这些值则忽略
	IgnoreErrs []error
	// 如果观察阶段返回错误是否在打印错误后退出
	ExitOnErrOfObserve bool
	// 启动服务函数
	RunServiceFn func() error
}

func WaitRun(app core.IApp, opt *WaitRunOption) error {
	if opt.ServiceType == "" {
		app.Fatal("ServiceType must not empty")
	}

	errChan := make(chan error, 1)
	go func(errChan chan error) {
		errChan <- opt.RunServiceFn()
	}(errChan)

	wait := time.NewTimer(time.Duration(app.GetConfig().Config().Frame.WaitServiceRunTime) * time.Millisecond) // 等待启动提前返回
	select {
	case <-wait.C:
	case <-app.BaseContext().Done():
		wait.Stop()
		return nil
	case err := <-errChan:
		wait.Stop()
		for _, e := range opt.IgnoreErrs {
			if err == e {
				return nil
			}
		}
		return err
	}

	// 开始等待服务启动阶段2
	go func(errChan chan error) {
		wait = time.NewTimer(time.Duration(app.GetConfig().Config().Frame.ServiceUnstableObserveTime) * time.Millisecond)
		select {
		case <-wait.C:
		case <-app.BaseContext().Done():
			wait.Stop()
		case err := <-errChan:
			wait.Stop()
			if err == nil {
				return
			}
			for _, e := range opt.IgnoreErrs {
				if err == e {
					return
				}
			}

			if opt.ExitOnErrOfObserve {
				app.Fatal("服务在观察阶段检测到错误", zap.String("serviceType", string(opt.ServiceType)), zap.Error(err))
			} else {
				app.Error("服务在观察阶段检测到错误", zap.String("serviceType", string(opt.ServiceType)), zap.Error(err))
			}
		}
	}(errChan)

	return nil
}
