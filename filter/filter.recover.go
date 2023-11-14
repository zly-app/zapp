package filter

import (
	"context"

	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/pkg/utils"
)

func init() {
	RegisterFilterCreator("base.recover", newRecoverFilter, newRecoverFilter)
}

var defRecoverFilter core.Filter = recoverFilter{}

func newRecoverFilter() core.Filter {
	return defRecoverFilter
}

type recoverFilter struct{}

func (r recoverFilter) Init(app core.IApp) error { return nil }

func (r recoverFilter) HandleInject(ctx context.Context, req, rsp interface{}, next core.FilterInjectFunc) error {
	err := utils.Recover.WrapCall(func() error {
		return next(ctx, req, rsp)
	})
	if utils.Recover.IsRecoverError(err) {
		GetCallMeta(ctx).SetPanic()
	}
	return err
}

func (r recoverFilter) Handle(ctx context.Context, req interface{}, next core.FilterFunc) (rsp interface{}, err error) {
	err = utils.Recover.WrapCall(func() error {
		rsp, err = next(ctx, req)
		return err
	})
	if utils.Recover.IsRecoverError(err) {
		GetCallMeta(ctx).SetPanic()
	}
	return
}

func (r recoverFilter) Close() error { return nil }
