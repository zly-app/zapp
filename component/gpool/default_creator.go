package gpool

import (
	"github.com/zly-app/zapp/core"
)

var defCreator = NewCreator()

func GetGPool(name string) core.IGPool {
	return defCreator.GetGPool(name)
}

func GetDefGPool() core.IGPool { return defCreator.GetGPool() }

func GetCreator() core.IGPools {
	return defCreator
}
