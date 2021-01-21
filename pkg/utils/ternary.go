/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2020/12/11
   Description :
-------------------------------------------------
*/

package utils

var Ternary = &ternaryUtil{}

type ternaryUtil struct{}

// 如果boole为真返回v1否则返回v2
func (*ternaryUtil) Ternary(boole bool, v1 interface{}, v2 interface{}) interface{} {
	if boole {
		return v1
	}
	return v2
}

// 顺序判断传入的数据, 如果某个数据不是其数据类型的零值则返回它, 否则返回最后一个数据
func (*ternaryUtil) Or(values ...interface{}) interface{} {
	var v interface{}
	for _, v = range values {
		if !Reflect.IsZero(v) {
			return v
		}
	}
	return v
}
