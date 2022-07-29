package compactor

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

var testData = `{"t":1650110771799311,"level":"debug","msg":"app初始化"}
{"t":1650110771811669,"level":"debug","msg":"app初始化完毕"}
{"t":1650110771811669,"level":"debug","msg":"启动app"}
{"t":1650110771811669,"level":"debug","msg":"启动插件"}
{"t":1650110771811669,"level":"debug","msg":"启动服务"}
{"t":1650110771811669,"level":"info","msg":"app已启动"}
{"t":1650110775760626,"level":"debug","msg":"app准备退出"}
{"t":1650110775760626,"level":"debug","msg":"关闭服务"}
{"t":1650110775760626,"level":"debug","msg":"关闭插件"}
{"t":1650110775760626,"level":"debug","msg":"释放组件资源"}
{"t":1650110775760626,"level":"debug","msg":"app已退出"}`

func TestRaw(t *testing.T) {
	r := bytes.NewBufferString(testData)
	w := bytes.NewBuffer(nil)
	c := NewRawCompactor()
	err := c.Compress(r, w)
	require.Nil(t, err)
	t.Log(len(testData), ">>", w.Len())

	w2 := bytes.NewBuffer(nil)
	err = c.UnCompress(w, w2)
	require.Nil(t, err)
	require.Equal(t, testData, w2.String())
}

func TestRawBytes(t *testing.T) {
	in := []byte(testData)
	c := NewRawCompactor()
	temp, err := c.CompressBytes(in)
	require.Nil(t, err)
	t.Log(len(testData), ">>", len(temp))

	in2, err := c.UnCompressBytes(temp)
	require.Nil(t, err)
	require.Equal(t, in, in2)
}
