/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/1/10
   Description :
-------------------------------------------------
*/

package logger

type ColorType byte

const (
	defaultColor ColorType = iota + '0' // 默认
	redColor                            // 红
	greenColor                          // 绿
	yellowColor                         // 黄
	blueColor                           // 蓝
	magentaColor                        // 紫
	cyanColor                           // 深绿
	writeColor                          // 灰色
)

func makeColorText(color ColorType, text string) string {
	if color == defaultColor {
		return text
	}
	bs := make([]byte, 9+len(text))
	copy(bs[:5], "\x1b[30m")
	bs[3] = byte(color)
	copy(bs[5:len(bs)-4], text)
	copy(bs[len(bs)-4:], "\x1b[0m")
	return string(bs)
}
