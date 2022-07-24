package compactor

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGzip(t *testing.T) {
	data := `{"t":1650110771799311,"level":"debug","msg":"app初始化"}`
	data += `{"t":1650110771811669,"level":"debug","msg":"app初始化完毕"}`
	data += `{"t":1650110771811669,"level":"debug","msg":"启动app"}`
	data += `{"t":1650110771811669,"level":"debug","msg":"启动插件"}`
	data += `{"t":1650110771811669,"level":"debug","msg":"启动服务"}`
	data += `{"t":1650110771811669,"level":"info","msg":"app已启动"}`
	data += `{"t":1650110775760626,"level":"debug","msg":"app准备退出"}`
	data += `{"t":1650110775760626,"level":"debug","msg":"关闭服务"}`
	data += `{"t":1650110775760626,"level":"debug","msg":"关闭插件"}`
	data += `{"t":1650110775760626,"level":"debug","msg":"释放组件资源"}`
	data += `{"t":1650110775760626,"level":"debug","msg":"app已退出"}`

	r := bytes.NewBufferString(data)
	w := bytes.NewBuffer(nil)
	c := NewGzipCompactor()
	err := c.Compress(r, w)
	require.Nil(t, err)
	t.Log(len(data), ">>", w.Len())

	w2 := bytes.NewBuffer(nil)
	err = c.UnCompress(w, w2)
	require.Nil(t, err)
	require.Equal(t, data, w2.String())
}
