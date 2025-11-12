package compactor

import (
	"io"

	"go.uber.org/zap"

	"github.com/zly-app/zapp/log"
)

type ICompactor interface {
	// 压缩
	Compress(in io.Reader, out io.Writer) error
	// 压缩
	CompressBytes(in []byte) (out []byte, err error)
	// 解压缩
	UnCompress(in io.Reader, out io.Writer) error
	// 解压缩
	UnCompressBytes(in []byte) (out []byte, err error)
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
			log.Log.Panic("Compactor重复注册", zap.String("name", name))
		}
	}
	compactorList[name] = c
}

// 获取压缩器, 压缩器不存在会panic
func GetCompactor(name string) ICompactor {
	c, ok := compactorList[name]
	if !ok {
		log.Log.Panic("未定义的CompactorName", zap.String("name", name))
	}
	return c
}

// 尝试获取压缩器
func TryGetCompactor(name string) (ICompactor, bool) {
	c, ok := compactorList[name]
	return c, ok
}
