package compactor

import (
	"bytes"
	"compress/gzip"
	"io"
)

const GzipCompactorName = "gzip"

type GzipCompactor struct{}

func (r GzipCompactor) Compress(in io.Reader, out io.Writer) error {
	w := gzip.NewWriter(out)
	_, err := io.Copy(w, in)
	if err != nil {
		return err
	}
	return w.Close()
}

func (r GzipCompactor) CompressBytes(in []byte) (out []byte, err error) {
	var outIO bytes.Buffer
	err = r.Compress(bytes.NewReader(in), &outIO)
	if err != nil {
		return nil, err
	}
	return outIO.Bytes(), nil
}

func (r GzipCompactor) UnCompress(in io.Reader, out io.Writer) error {
	read, err := gzip.NewReader(in)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, read)
	if err != nil {
		return err
	}
	return read.Close()
}

func (r GzipCompactor) UnCompressBytes(in []byte) (out []byte, err error) {
	var outIO bytes.Buffer
	err = r.UnCompress(bytes.NewReader(in), &outIO)
	if err != nil {
		return nil, err
	}
	return outIO.Bytes(), nil
}

func NewGzipCompactor() ICompactor {
	return GzipCompactor{}
}
