package config

import (
	"sync/atomic"
)

type IUserWhiteList interface {
	// 数据构建
	make()
	// 查询是否在白名单中
	IsWhiteList(uid string) bool
}

var _ IUserWhiteList = (*UserWhiteList)(nil)

// 用户白名单, 多个数据同时存在时只要复合任意一个就行
type UserWhiteList struct {
	Uids    []string // 指定的用户
	Percent uint8    // 灰度比例, 百分比
	Tails   []string // 用户后两位尾号

	uidMap  map[string]struct{}
	tailMap map[string]struct{}
	incrV   uint32
}

func (u *UserWhiteList) make() {
	if u == nil {
		return
	}

	u.uidMap = make(map[string]struct{}, len(u.Uids))
	for _, uid := range u.Uids {
		u.uidMap[uid] = struct{}{}
	}

	u.tailMap = make(map[string]struct{}, len(u.Tails))
	for _, tail := range u.Tails {
		u.tailMap[tail] = struct{}{}
	}
}

func (u *UserWhiteList) IsWhiteList(uid string) bool {
	if u == nil {
		return false
	}

	if u.Percent >= 100 {
		return true
	}

	if _, ok := u.uidMap[uid]; ok {
		return true
	}

	if len(uid) >= 2 {
		tail := uid[len(uid)-2:]
		if _, ok := u.tailMap[tail]; ok {
			return true
		}
	}

	if int32(u.Percent) <= 0 {
		return false
	}

	// 伪随机(轮询)
	if atomic.AddUint32(&u.incrV, 1)%100 < uint32(u.Percent) {
		return true
	}

	return false
}

func init() {
	AddResetInjectObjCallback[IUserWhiteList](func(obj IUserWhiteList, isField bool) {
		// 构建用户白名单数据
		if u, ok := obj.(IUserWhiteList); ok {
			u.make()
		}
	})
}
