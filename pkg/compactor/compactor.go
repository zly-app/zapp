package compactor

import (
	"io"

	"go.uber.org/zap"

	"github.com/zly-app/zapp/logger"
)

type ICompactor interface {
	// 压缩
	Compress(in io.Reader, out io.Writer) error
	// 解压缩
	UnCompress(in io.Reader, out io.Writer) error
}

var compactorList = map[string]ICompactor{
	RawCompactorName:  NewRawCompactor(),
	ZStdCompactorName: NewZStdCompactor(),
	GzipCompactorName: NewGzipCompactor(),
}

// 注册压缩器, 重复注册会panic
func RegistryCompactor(name string, c ICompactor, replace ...bool) {
	if len(replace) == 0 || !replace[0] {
		if _, ok := compactorList[name]; ok {
			logger.Log.Panic("Compactor重复注册", zap.String("name", name))
		}
	}
	compactorList[name] = c
}

// 获取压缩器, 压缩器不存在会panic
func GetCompactor(name string) ICompactor {
	c, ok := compactorList[name]
	if !ok {
		logger.Log.Panic("未定义的CompactorName", zap.String("name", name))
	}
	return c
}

// 尝试获取压缩器
func TryGetCompactor(name string) (ICompactor, bool) {
	c, ok := compactorList[name]
	return c, ok
}
