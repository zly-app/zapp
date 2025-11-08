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

func (*timeoutFilter) Name() string { return "base.timeout" }

func (r *timeoutFilter) Init(app core.IApp) error {
	err := config.Conf.ParseFilterConfig("base.timeout", r, true)
	if err != nil {
		return err
	}
	return nil
}

func (r *timeoutFilter) getClientTimeout(clientType, clientName string) int64 {
	ct, ok := r.Client[clientType]
	if ok {
		l, ok := ct[clientName]
		if ok {
			return l
		}
		l, ok = ct[defName] // 默认客户端组件
		if ok {
			return l
		}
	}

	ct, ok = r.Client[defName]
	if ok {
		l, ok := ct[defName]
		if ok {
			return l
		}
	}
	return defTimeout
}
func (r *timeoutFilter) getServiceTimeout(serviceName string) int64 {
	l, ok := r.Service[serviceName]
	if ok {
		return l
	}
	l, ok = r.Service[defName]
	if ok {
		return l
	}
	return defTimeout
}
func (r *timeoutFilter) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	meta := GetCallMeta(ctx)
	t := defTimeout
	if meta.IsClientMeta() {
		t = r.getClientTimeout(meta.ClientType(), meta.ClientName())
	} else if meta.IsServiceMeta() {
		t = r.getServiceTimeout(meta.ServiceName())
	}

	if t <= 0 {
		return ctx, func() {}
	}

	deadline, ok := ctx.Deadline()
	if ok && time.Now().Add(time.Millisecond*time.Duration(t)).After(deadline) { // 如果设置的超时时间在当前截止时间之后, 则设置超时时间是无意义的
		return ctx, func() {}
	}

	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*time.Duration(t))
	return ctx, cancel
}

func (r *timeoutFilter) HandleInject(ctx context.Context, req, rsp interface{}, next core.FilterInjectFunc) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	return next(ctx, req, rsp)
}

func (r *timeoutFilter) Handle(ctx context.Context, req interface{}, next core.FilterFunc) (rsp interface{}, err error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	return next(ctx, req)
}

func (r *timeoutFilter) Close() error { return nil }
