package monitor

import (
	"sync"
)

const (
	// SelfMonitorKey 自身状态监控的 KEY
	SelfMonitorKey = "__WYJ_RecordMonitorSpecKeyTemplateException"
)

// 监控数据存储核心及方法

// Storage 监控指标的存储系统
type Storage struct {
	sync.RWMutex
	MetricMap *MetricNameMap // 指标的映射

	NowMonitor           *OneMinStorage   // 当前监控数据
	HistoryMonitor       []*OneMinStorage //历史的监控数据
	HistoryVersionNumber int              // 历史版本数
	Cursor               int              // 历史版本游标
}

// NewStorage 初始化一个核心的存储
func NewStorage(history int) *Storage {
	s := &Storage{
		MetricMap:            NewMetricNameMap(),
		NowMonitor:           nil,
		HistoryMonitor:       make([]*OneMinStorage, history+1),
		HistoryVersionNumber: history,
		Cursor:               0,
	}

	s.NowMonitor = NewOneMinStorage()

	return s
}

// nextSpecValue 生成新的模版
func (s *Storage) nextSpecValue() map[int]*SpecValue {
	s.MetricMap.RLock()
	template := make(map[int]*SpecValue, len(s.MetricMap.Map))

	for k, metric := range s.MetricMap.Map {

		v, err := initSpecValue(metric.Type)
		if err != nil {
			Logger.Printf("Init SpecValue Error : %s", err)
			continue
		}

		template[k] = v
	}

	s.MetricMap.RUnlock()
	return template
}

// NextMonitor 切换监控版本数据, 迭代下一版数据
// 返回当前版本的数据
func (s *Storage) NextMonitor() (now *OneMinStorage) {
	// 初始化 SpecValue
	next := NewOneMinStorage()
	next.PersistentData = s.nextSpecValue()

	s.Lock()
	// 记录当前的监控数据,并返回交友切换代码做后续 上传 or 落地
	now = s.NowMonitor
	// 切换 及 判断
	s.NowMonitor = next

	// 数据添加到历史版本中,并移动游标
	s.HistoryMonitor[s.Cursor] = now
	if s.Cursor == s.HistoryVersionNumber-1 {
		s.Cursor = -1
	}
	s.Cursor++

	s.Unlock()

	return
}
