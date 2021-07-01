package msgbus_consume

import (
	"sync"

	"github.com/zly-app/zapp/core"
)

type MsgbusConsumeService struct {
	app              core.IApp
	subscribes       map[string][]uint32 // 订阅者
	globalSubscribes []uint32            // 全局订阅者
	mx               sync.Mutex          // 锁 subscribes,globalSubscribes
}

func (m *MsgbusConsumeService) Inject(a ...interface{}) {
	for _, v := range a {
		conf, ok := v.(*ConsumerConfig)
		if !ok {
			m.app.Fatal("msgbus消费服务注入类型错误, 它必须能转为 *msgbus_consume.ConsumerConfig")
		}

		m.mx.Lock()
		if conf.IsGlobal {
			subId := m.app.GetComponent().SubscribeGlobal(conf.ThreadCount, conf.Handler)
			m.globalSubscribes = append(m.globalSubscribes, subId)
		} else {
			subId := m.app.GetComponent().Subscribe(conf.Topic, conf.ThreadCount, conf.Handler)
			m.subscribes[conf.Topic] = append(m.subscribes[conf.Topic], subId)
		}
		m.mx.Unlock()
	}
}

func (m *MsgbusConsumeService) Start() error {
	return nil
}
func (m *MsgbusConsumeService) Close() error {
	c := m.app.GetComponent()

	m.mx.Lock()
	// 取消订阅主题
	for topic, subscribes := range m.subscribes {
		for _, subId := range subscribes {
			c.Unsubscribe(topic, subId)
		}
	}
	m.subscribes = make(map[string][]uint32)

	// 取消订阅全局
	for _, subId := range m.globalSubscribes {
		c.UnsubscribeGlobal(subId)
	}
	m.globalSubscribes = make([]uint32, 0)

	m.mx.Unlock()
	return nil
}

func NewMsgbusConsumeService(app core.IApp) core.IService {
	// 加载配置
	return &MsgbusConsumeService{
		app:              app,
		subscribes:       make(map[string][]uint32),
		globalSubscribes: make([]uint32, 0),
	}
}
