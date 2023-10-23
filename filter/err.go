package filter

import (
	"context"
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
	case context.DeadlineExceeded, context.Canceled:
		return -1, CodeTypeTimeoutOrCancel, err
	case context.Canceled:
		return -1, CodeTypeTimeoutOrCancel, err
	}

	meta := GetCallMeta(ctx)
	if meta.HasPanic() {
		return -2, CodeTypeException, err
	}

	return -3, CodeTypeFail, err
}
