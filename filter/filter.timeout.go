package filter

import (
	"context"
	"time"

	"github.com/zly-app/zapp/config"
	"github.com/zly-app/zapp/core"
)

const defTimeout int64 = 5000

func init() {
	RegisterFilterCreator("base.timeout", newTimeoutFilter, newTimeoutFilter)
}

var defTimeoutFilter core.Filter = &timeoutFilter{
	Client:  make(map[string]map[string]int64),
	Service: make(map[string]int64),
}

func newTimeoutFilter() core.Filter {
	return defTimeoutFilter
}

type timeoutFilter struct {
	Client  map[string]map[string]int64
	Service map[string]int64
}

func (r *timeoutFilter) Init() error {
	err := config.Conf.ParseFilterConfig("base.timeout", r, true)
	if err != nil {
		return err
	}
	return nil
}

func (r *timeoutFilter) getClientTimeout(clientType, clientName string) int64 {
	ct, ok := r.Client[clientType]
	if !ok { // 没有找到 clientType 则用全局默认
		ct, ok = r.Client[defName]
		if ok {
			return ct[defName]
		}
		return defTimeout
	}

	t, ok := ct[clientName]
	if ok {
		return t
	}
	return ct[defName]
}
func (r *timeoutFilter) getServiceTimeout(serviceName string) int64 {
	t, ok := r.Service[serviceName]
	if ok {
		return t
	}
	return r.Service[defName]
}
func (r *timeoutFilter) getTimeout(ctx context.Context) int64 {
	meta := GetCallMeta(ctx)
	t := defTimeout
	if meta.IsClientMeta() {
		t = r.getClientTimeout(meta.ClientType(), meta.ClientName())
	} else if meta.IsServiceMeta() {
		t = r.getServiceTimeout(meta.ServiceName())
	}
	return t
}

func (r *timeoutFilter) HandleInject(ctx context.Context, req, rsp interface{}, next core.FilterInjectFunc) error {
	t := r.getTimeout(ctx)
	if t > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Millisecond*time.Duration(t))
		defer cancel()
	}

	return next(ctx, req, rsp)
}

func (r *timeoutFilter) Handle(ctx context.Context, req interface{}, next core.FilterFunc) (rsp interface{}, err error) {
	t := r.getTimeout(ctx)
	if t > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Millisecond*time.Duration(t))
		defer cancel()
	}

	return next(ctx, req)
}

func (r *timeoutFilter) Close() error { return nil }
