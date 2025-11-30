package filter

import (
	"context"

	"github.com/zly-app/zapp/pkg/utils"
)

const (
	CodeTypeSuccess         = "success"
	CodeTypeTimeoutOrCancel = "timeoutOrCancel"
	CodeTypeFail            = "fail"
	CodeTypeException       = "exception"
)

type GetErrCodeFunc func(ctx context.Context, rsp interface{}, err error) (code int, codeType string, replaceErr error)

var DefaultGetErrCodeFunc GetErrCodeFunc = func(ctx context.Context, rsp interface{}, err error) (
	code int, codeType string, replaceErr error) {
	if err == nil {
		return 0, CodeTypeSuccess, nil
	}

	switch err {
	case context.DeadlineExceeded:
		return -1, CodeTypeTimeoutOrCancel, err
	}

	meta := GetCallMeta(ctx)
	if meta.HasPanic() {
		return -2, CodeTypeException, err
	}
	if utils.Recover.IsRecoverError(err) {
		return -2, CodeTypeException, err
	}

	return -3, CodeTypeFail, err
}
