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
添加当watch对象发生变化时的回调

如果watch对象实现了T, 则会调用回调函数并传入watch对象的值, 其isField为false
如果watch对象是一个struct, 则会遍历其导出字段(包括继承)判断实现了 T, 则会调用回调函数并传入其字段的值, 其isField为true.
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

	ignoreAnonymousMark := make(map[reflect.Type]struct{}) // 如果在外层处理了, 那么当内部为匿名字段时要忽略处理
	match := func(matchType func(t reflect.Type) bool, isField, isAnonymous bool, get func() interface{}) {
		for rt, callback := range resetInjectObjCallbacks {
			if !matchType(rt) {
				continue
			}
			if !isField {
				ignoreAnonymousMark[rt] = struct{}{} // 标记已被外层处理
			} else if isAnonymous {
				if _, ok := ignoreAnonymousMark[rt]; ok { // 已被外层处理过了
					continue
				}
			}
			callback(get(), isField)
		}
	}

	rv := reflect.ValueOf(injectObj)
	if rv.Type().Kind() == reflect.Ptr {
		match(rv.Type().AssignableTo, false, false, func() interface{} {
			return rv.Interface()
		})
	}

	rv = reflect.Indirect(rv)
	rt := rv.Type()

	if rv.Kind() == reflect.Struct {
		num := rt.NumField()
		for i := 0; i < num; i++ {
			ft := rt.Field(i)
			fv := rv.Field(i)
			if !fv.CanAddr() || ft.PkgPath != "" { // 无法获取值 || 未导出
				continue
			}

			isPtr := fv.Kind() == reflect.Ptr
			if !isPtr {
				match(fv.Addr().CanConvert, true, ft.Anonymous, func() interface{} {
					return fv.Addr().Interface()
				})
			}
			match(fv.CanConvert, true, ft.Anonymous, func() interface{} {
				return fv.Interface()
			})
		}
	}
	return
}
