package compactor

import (
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

func NewGzipCompactor() ICompactor {
	return GzipCompactor{}
}
