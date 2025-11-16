package filter

import (
	"net/http"

	"github.com/bytedance/sonic"
)

const callerMetaHeaderKey = "X-Caller-Meta"

// 将主调信息写入到headers中
func SaveCallerMeta2Header(headers http.Header, callerMeta CallerMeta) {
	text, _ := sonic.MarshalString(callerMeta)
	headers.Add(callerMetaHeaderKey, text)
}

// 从headers中获取主调信息
func GetCallerMetaByHeader(header http.Header) CallerMeta {
	vs := header.Get(callerMetaHeaderKey)
	var callerMeta CallerMeta
	_ = sonic.UnmarshalString(vs, &callerMeta)
	return callerMeta
}
