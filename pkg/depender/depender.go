package depender

import (
	"fmt"
)

type Item interface {
	Name() string
	DependsOn() []string
	Start() error
	Close()
}

type Depender interface {
	Start() error
	Close()
}

type DependerCli struct {
	items       []Item              // 所有项
	readyItem   map[string]struct{} // 已启动项
	startedItem []Item              // 已启动项的启动顺序
}

func NewDepender(items []Item) Depender {
	d := &DependerCli{
		items:       items,
		readyItem:   make(map[string]struct{}, len(items)),
		startedItem: make([]Item, 0, len(items)),
	}
	return d
}

// 启动
func (d *DependerCli) Start() error {
	forNums := 0
	maxForNums := len(d.items) * 10 // 防止出现循环依赖

	ch := make(chan Item, len(d.items))
	for i := range d.items {
		ch <- d.items[i]
	}

	for len(d.startedItem) < len(d.items) {
		item := <-ch

		forNums++
		if forNums > maxForNums {
			return fmt.Errorf("There may be cyclic dependencies, item=%s", item.Name())
		}

		// 检查依赖
		ready := true
		for _, dep := range item.DependsOn() {
			_, ok := d.readyItem[dep]
			if !ok {
				ready = false
				break
			}
		}
		if !ready {
			ch <- item
			continue
		}

		err := item.Start()
		if err != nil {
			return fmt.Errorf("start err. item=%s, err=%v", item.Name(), err)
		}

		d.readyItem[item.Name()] = struct{}{}
		d.startedItem = append(d.startedItem, item)
	}
	return nil
}

// 关闭
func (d *DependerCli) Close() {
	for i := len(d.startedItem) - 1; i >= 0; i-- {
		item := d.startedItem[i]
		item.Close()
	}
}

type itemCli struct {
	name      string
	dependsOn []string
	start     func() error
	close     func()
}

func (i itemCli) Name() string        { return i.name }
func (i itemCli) DependsOn() []string { return i.dependsOn }
func (i itemCli) Start() error        { return i.start() }
func (i itemCli) Close()              { i.close() }

func NewItem(name string, dependsOn []string, start func() error, close func()) Item {
	return itemCli{name, dependsOn, start, close}
}
