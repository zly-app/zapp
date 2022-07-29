package compactor

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGzip(t *testing.T) {
	r := bytes.NewBufferString(testData)
	w := bytes.NewBuffer(nil)
	c := NewGzipCompactor()
	err := c.Compress(r, w)
	require.Nil(t, err)
	t.Log(len(testData), ">>", w.Len())

	w2 := bytes.NewBuffer(nil)
	err = c.UnCompress(w, w2)
	require.Nil(t, err)
	require.Equal(t, testData, w2.String())
}

func TestGzipBytes(t *testing.T) {
	in := []byte(testData)
	c := NewGzipCompactor()
	temp, err := c.CompressBytes(in)
	require.Nil(t, err)
	t.Log(len(testData), ">>", len(temp))

	in2, err := c.UnCompressBytes(temp)
	require.Nil(t, err)
	require.Equal(t, in, in2)
}
