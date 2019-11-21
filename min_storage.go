package monitor

import (
	"sync"
	"time"
)

// 一分钟的存储定义及方法

// OneMinStorage 一分钟的指标存储
// Ts 先占位, 之后可能会有需要返回其创建是的分钟或其它
type OneMinStorage struct {
	sync.RWMutex
	Ts             time.Time          // 初始化时候时间
	PersistentData map[int]*SpecValue // 监控数据
	Data           map[string]float64 // 其它监控
}

// NewOneMinStorage 初始化一个一分钟的存储
func NewOneMinStorage() *OneMinStorage {
	return &OneMinStorage{
		Ts:             time.Now(),
		PersistentData: make(map[int]*SpecValue),
		Data:           make(map[string]float64),
	}
}

// Add 针对一个具体指标名添加一个 float64 的值
func (oms *OneMinStorage) Add(name string, value float64) {
	oms.Lock()
	oms.Data[name] += value
	oms.Unlock()
}

// AddPersistent 针对一个持久化的指标名添加一个 float64 的值
func (oms *OneMinStorage) AddPersistent(MapID int, metricType int, value float64) {
	oms.Lock()
	defer oms.Unlock()

	if sv, ok := oms.PersistentData[MapID]; ok {
		sv.Lock()

		switch metricType {
		case BaseMetric:
			sv.Sum += value
		case SumMetric:
			sv.Sum += value
		case AvgMetric:
			sv.Sum += value
			sv.Count++
		case CountMetric:
			sv.Count++
		case CountSumMetric:
			sv.Sum += value
			sv.Count++
		case CountAvgMetric:
			sv.Sum += value
			sv.Count++
		case QuantileMetric:
			sv.Otd.Add(value)
		}

		sv.Unlock()
		return
	}

	Logger.Printf("AddPersistent Not Have MapID: %d, func break", MapID)
}

// Set 针对一个具体指标名添加一个 float64 的值
// 暂时用锁实现
func (oms *OneMinStorage) Set(name string, value float64) {
	oms.Lock()
	oms.Data[name] = value
	oms.Unlock()
}

// SetPersistent 针对一个持久化的指标名添加一个 float64 的值
// 分位数不支持 Set 方法
// 基础的和Sum类型则直接使用 value 值
// Avg 方法使用 value 方法同时将 Count 置为 1
// Count 忽略 value, 直接将 Count 类型置为 1
func (oms *OneMinStorage) SetPersistent(MapID int, metricType int, value float64) {
	oms.Lock()
	defer oms.Unlock()

	if sv, ok := oms.PersistentData[MapID]; ok {
		sv.Lock()

		switch metricType {
		case BaseMetric:
			sv.Sum = value
		case SumMetric:
			sv.Sum = value
		case CountMetric:
			sv.Count = 1
		case AvgMetric:
			sv.Sum = value
			sv.Count = 1
		case CountSumMetric:
			sv.Sum = value
			sv.Count = 1
		case CountAvgMetric:
			sv.Sum = value
			sv.Count = 1
		case QuantileMetric:
			Logger.Println("QuantileMetric Metric Cant Use Set Method")
		}

		sv.Unlock()
		return
	}

	Logger.Printf("Set Not Have MapID: %d, func break", MapID)
}

// Get 获取某个监控指标的 value
func (oms *OneMinStorage) Get(name string) (value float64, err error) {
	oms.RLock()

	ok := false
	if value, ok = oms.Data[name]; ok {
		err = nil
	} else {
		err = &ErrNotFoundMetric{ID: -1, Name: name}
	}
	oms.RUnlock()
	return
}

// GetPersistent 获取某个持久化的监控指标的 value
func (oms *OneMinStorage) GetPersistent(MapID int) *SpecValue {
	return oms.PersistentData[MapID]
}

// Len 获取当前分钟有多少个监控指标
func (oms *OneMinStorage) Len() int {
	return len(oms.Data) + len(oms.PersistentData)
}

// GetAll 用在落地 or 上传的时候 TODO
// 可能需要在给出对应的映射关系, 先占位
func (oms *OneMinStorage) GetAll() {
}
