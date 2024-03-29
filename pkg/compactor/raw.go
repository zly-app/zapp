package compactor

import (
	"io"
)

const RawCompactorName = "raw"

// 原始数据, 不进行任何压缩
type RawCompactor struct{}

func (r RawCompactor) Compress(in io.Reader, out io.Writer) error {
	_, err := io.Copy(out, in)
	return err
}

func (r RawCompactor) CompressBytes(in []byte) (out []byte, err error) {
	return in, nil
}

func (r RawCompactor) UnCompress(in io.Reader, out io.Writer) error {
	_, err := io.Copy(out, in)
	return err
}

func (r RawCompactor) UnCompressBytes(in []byte) (out []byte, err error) {
	return in, nil
}

func NewRawCompactor() ICompactor {
	return RawCompactor{}
}
