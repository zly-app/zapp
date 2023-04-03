package config

import (
	"reflect"
)

func init() {
	AddResetInjectStructuredHook(resetInjectObjCallback)
}

type ResetInjectStructuredHook func(injectObj interface{}, newData []byte) (result interface{}, resultData []byte, cancel bool)

var resetInjectStructuredHooks []ResetInjectStructuredHook

// 添加重设注入结构拦截
func AddResetInjectStructuredHook(hooks ...ResetInjectStructuredHook) {
	resetInjectStructuredHooks = append(resetInjectStructuredHooks, hooks...)
}

var resetInjectObjCallbacks = make(map[reflect.Type]func(injectObj interface{}, isField bool))

/*
添加重设结构注入结构回调, 只有注入结构是指针才有效

如果注入结构实现了 T, 则会调用回调函数并传入注入结构的值, 其isField为false
如果注入结构是一个struct, 则会遍历其导出字段(包括继承)判断实现了 T , 则会调用回调函数并传入其字段的值, 其isField为true.
*/
func AddResetInjectObjCallback[T any](callback func(obj T, isField bool)) {
	o := new(T)
	rt := reflect.TypeOf(o).Elem()
	resetInjectObjCallbacks[rt] = func(obj interface{}, isField bool) {
		callback(obj.(T), isField)
	}
}

func resetInjectObjCallback(injectObj interface{}, newData []byte) (result interface{}, resultData []byte, cancel bool) {
	result = injectObj
	resultData = newData
	cancel = false

	match := func(t reflect.Type, isField bool, get func() interface{}) {
		for rt, callback := range resetInjectObjCallbacks {
			if t.AssignableTo(rt) {
				callback(get(), isField)
			}
		}
	}

	rv := reflect.ValueOf(injectObj)
	if rv.Type().Kind() == reflect.Ptr {
		match(rv.Type(), false, func() interface{} {
			return rv.Interface()
		})
	}

	rv = reflect.Indirect(rv)
	rt := rv.Type()

	if rv.Kind() == reflect.Struct {
		num := rt.NumField()
		for i := 0; i < num; i++ {
			ft := rt.Field(i)
			if ft.PkgPath != "" || ft.Type.Kind() != reflect.Ptr {
				continue
			}
			match(ft.Type, true, func() interface{} {
				return rv.Field(i).Interface()
			})
		}
	}
	return
}
