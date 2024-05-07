package config

import (
	"sync/atomic"
)

type IUserMatcher interface {
	// 数据构建
	make()
	// 查询是否命中
	IsHit(uid string) bool
}

var _ IUserMatcher = (*UserMatcher)(nil)

// 用户匹配器, 多个数据同时存在时只要复合任意一个就行
type UserMatcher struct {
	Uids    []string // 指定的用户
	Percent uint8    // 灰度比例, 百分比
	Tails   []string // 用户后两位尾号

	uidMap  map[string]struct{}
	tailMap map[string]struct{}
	incrV   uint32
}

func (u *UserMatcher) make() {
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

func (u *UserMatcher) IsHit(uid string) bool {
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
	AddResetInjectObjCallback[IUserMatcher](func(obj IUserMatcher, isField bool) {
		if u, ok := obj.(IUserMatcher); ok {
			u.make()
		}
	})
}
